//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package logicrunner - infrastructure for executing smartcontracts
package logicrunner

import (
	"bytes"
	"context"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/ThreeDotsLabs/watermill"
	watermillMsg "github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/infrastructure/gochannel"

	"github.com/insolar/insolar/log"

	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/flow/dispatcher"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
	"github.com/insolar/insolar/logicrunner/artifacts"
	"github.com/insolar/insolar/logicrunner/builtin"
	lrCommon "github.com/insolar/insolar/logicrunner/common"
	"github.com/insolar/insolar/logicrunner/goplugin"
)

const maxQueueLength = 10

type Ref = insolar.Reference

// Context of one contract execution
type ObjectState struct {
	sync.Mutex

	ExecutionState *ExecutionState
	Validation     *ExecutionState
}

type Error struct {
	Err      error
	Request  *Ref
	Contract *Ref
	Method   string
}

func (lre Error) Error() string {
	var buffer bytes.Buffer

	buffer.WriteString(lre.Err.Error())
	if lre.Contract != nil {
		buffer.WriteString(" Contract=" + lre.Contract.String())
	}
	if lre.Method != "" {
		buffer.WriteString(" Method=" + lre.Method)
	}
	if lre.Request != nil {
		buffer.WriteString(" Request=" + lre.Request.String())
	}

	return buffer.String()
}

func (st *ObjectState) GetModeState(mode insolar.CallMode) (rv *ExecutionState, err error) {
	switch mode {
	case insolar.ExecuteCallMode:
		rv = st.ExecutionState
	case insolar.ValidateCallMode:
		rv = st.Validation
	default:
		err = errors.Errorf("'%d' is unknown object processing mode", mode)
	}

	if rv == nil && err != nil {
		err = errors.Errorf("object is not in '%s' mode", mode)
	}
	return rv, err
}

func (st *ObjectState) MustModeState(mode insolar.CallMode) *ExecutionState {
	res, err := st.GetModeState(mode)
	if err != nil {
		panic(err)
	}
	if res.CurrentList.Empty() {
		panic("object " + res.Ref.String() + " has no Current")
	}
	return res
}

func (st *ObjectState) WrapError(err error, message string) error {
	if err == nil {
		err = errors.New(message)
	} else {
		err = errors.Wrap(err, message)
	}
	return Error{
		Err: err,
	}
}

func makeWMMessage(ctx context.Context, payLoad watermillMsg.Payload, msgType string) *watermillMsg.Message {
	wmMsg := watermillMsg.NewMessage(watermill.NewUUID(), payLoad)
	wmMsg.Metadata.Set(bus.MetaTraceID, inslogger.TraceID(ctx))
	wmMsg.Metadata.Set(bus.MetaType, msgType)

	return wmMsg
}

// LogicRunner is a general interface of contract executor
type LogicRunner struct {
	MessageBus                 insolar.MessageBus                 `inject:""`
	ContractRequester          insolar.ContractRequester          `inject:""`
	NodeNetwork                insolar.NodeNetwork                `inject:""`
	PlatformCryptographyScheme insolar.PlatformCryptographyScheme `inject:""`
	ParcelFactory              message.ParcelFactory              `inject:""`
	PulseAccessor              pulse.Accessor                     `inject:""`
	ArtifactManager            artifacts.Client                   `inject:""`
	DescriptorsCache           artifacts.DescriptorsCache         `inject:""`
	JetCoordinator             jet.Coordinator                    `inject:""`

	Executors    [insolar.MachineTypesLastID]insolar.MachineLogicExecutor
	machinePrefs []insolar.MachineType
	Cfg          *configuration.LogicRunner

	state      map[Ref]*ObjectState // if object exists, we are validating or executing it right now
	stateMutex sync.RWMutex

	rpc *lrCommon.RPC

	stopLock   sync.Mutex
	isStopping bool
	stopChan   chan struct{}

	// Inner dispatcher will be merged with FlowDispatcher after
	// complete migration to watermill.
	FlowDispatcher      *dispatcher.Dispatcher
	innerFlowDispatcher *dispatcher.Dispatcher
	publisher           watermillMsg.Publisher
	router              *watermillMsg.Router
}

// NewLogicRunner is constructor for LogicRunner
func NewLogicRunner(cfg *configuration.LogicRunner) (*LogicRunner, error) {
	if cfg == nil {
		return nil, errors.New("LogicRunner have nil configuration")
	}
	res := LogicRunner{
		Cfg:   cfg,
		state: make(map[Ref]*ObjectState),
	}
	res.rpc = lrCommon.NewRPC(NewRPCMethods(&res), cfg)

	return &res, nil
}

func InitHandlers(lr *LogicRunner, s bus.Sender) (*watermillMsg.Router, error) {
	wmLogger := log.NewWatermillLogAdapter(inslogger.FromContext(context.Background()))
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, wmLogger)

	dep := &Dependencies{
		Publisher: pubSub,
		lr:        lr,
		Sender:    s,
	}

	initHandle := func(msg *watermillMsg.Message) *Init {
		return &Init{
			dep: dep,
			// TODO: use wm.Message instead of bus.Message
			Message: msg,
		}
	}
	lr.FlowDispatcher = dispatcher.NewDispatcher(func(msg *watermillMsg.Message) flow.Handle {
		return initHandle(msg).Present
	}, func(msg *watermillMsg.Message) flow.Handle {
		return initHandle(msg).Future
	})

	innerInitHandle := func(msg *watermillMsg.Message) *InnerInit {
		return &InnerInit{
			dep:     dep,
			Message: msg,
		}
	}

	lr.innerFlowDispatcher = dispatcher.NewDispatcher(func(msg *watermillMsg.Message) flow.Handle {
		return innerInitHandle(msg).Present
	}, func(msg *watermillMsg.Message) flow.Handle {
		return innerInitHandle(msg).Present
	})

	router, err := watermillMsg.NewRouter(watermillMsg.RouterConfig{}, wmLogger)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating new watermill router")
	}

	router.AddNoPublisherHandler(
		"InnerMsgHandler",
		InnerMsgTopic,
		pubSub,
		lr.innerFlowDispatcher.InnerSubscriber,
	)
	go func() {
		if err := router.Run(); err != nil {
			ctx := context.Background()
			inslogger.FromContext(ctx).Error("Error while running router", err)
		}
	}()
	<-router.Running()

	lr.router = router
	lr.publisher = pubSub

	return router, nil
}

func (lr *LogicRunner) initializeBuiltin(_ context.Context) error {
	bi := builtin.NewBuiltIn(lr.MessageBus, lr.ArtifactManager)
	if err := lr.RegisterExecutor(insolar.MachineTypeBuiltin, bi); err != nil {
		return err
	}
	lr.machinePrefs = append(lr.machinePrefs, insolar.MachineTypeBuiltin)

	// TODO: insert all necessary descriptors here
	codeDescriptors := builtin.InitializeCodeDescriptors()
	for _, codeDescriptor := range codeDescriptors {
		lr.ArtifactManager.InjectCodeDescriptor(*codeDescriptor.Ref(), codeDescriptor)
	}

	prototypeDescriptors := builtin.InitializePrototypeDescriptors()
	for _, prototypeDescriptor := range prototypeDescriptors {
		lr.ArtifactManager.InjectObjectDescriptor(*prototypeDescriptor.HeadRef(), prototypeDescriptor)
	}

	lr.ArtifactManager.InjectFinish()

	lrCommon.CurrentProxyCtx = builtin.NewProxyHelper(NewRPCMethods(lr))

	return nil
}

func (lr *LogicRunner) initializeGoPlugin(ctx context.Context) error {
	if lr.Cfg.RPCListen != "" {
		lr.rpc.Start(ctx)
	}

	gp, err := goplugin.NewGoPlugin(lr.Cfg, lr.MessageBus, lr.ArtifactManager)
	if err != nil {
		return err
	}
	if err := lr.RegisterExecutor(insolar.MachineTypeGoPlugin, gp); err != nil {
		return err
	}
	lr.machinePrefs = append(lr.machinePrefs, insolar.MachineTypeGoPlugin)
	return nil
}

// Start starts logic runner component
func (lr *LogicRunner) Start(ctx context.Context) error {
	if lr.Cfg.BuiltIn != nil {
		log.Error("Initializing builtin")
		if err := lr.initializeBuiltin(ctx); err != nil {
			log.Errorf("Initializing builtin not done: %s", err.Error())
			return err
		}
		log.Error("Initializing builtin done")
	}

	if lr.Cfg.GoPlugin != nil {
		if err := lr.initializeGoPlugin(ctx); err != nil {
			return err
		}
	}

	lr.RegisterHandlers()

	return nil
}

func (lr *LogicRunner) RegisterHandlers() {
	lr.MessageBus.MustRegister(insolar.TypeValidateCaseBind, lr.HandleValidateCaseBindMessage)
	lr.MessageBus.MustRegister(insolar.TypeValidationResults, lr.HandleValidationResultsMessage)
}

// Stop stops logic runner component and its executors
func (lr *LogicRunner) Stop(ctx context.Context) error {
	reterr := error(nil)
	for _, e := range lr.Executors {
		if e == nil {
			continue
		}
		err := e.Stop()
		if err != nil {
			reterr = errors.Wrap(reterr, err.Error())
		}
	}

	if err := lr.rpc.Stop(ctx); err != nil {
		return err
	}
	if err := lr.router.Close(); err != nil {
		return err
	}

	return reterr
}

func (lr *LogicRunner) GracefulStop(ctx context.Context) error {
	inslogger.FromContext(ctx).Debug("LogicRunner.GracefulStop starts ...")

	lr.stopLock.Lock()
	if !lr.isStopping {
		lr.isStopping = true
		lr.stopChan = make(chan struct{}, 1)
	}
	lr.stopLock.Unlock()

	inslogger.FromContext(ctx).Debug("LogicRunner.GracefulStop wait ...")
	<-lr.stopChan
	inslogger.FromContext(ctx).Debug("LogicRunner.GracefulStop ends ...")
	return nil
}

func (lr *LogicRunner) CheckOurRole(ctx context.Context, msg insolar.Message, role insolar.DynamicRole) error {
	// TODO do map of supported objects for pulse, go to jetCoordinator only if map is empty for ref
	target := msg.DefaultTarget()
	isAuthorized, err := lr.JetCoordinator.IsAuthorized(
		ctx, role, *target.Record(), lr.pulse(ctx).PulseNumber, lr.JetCoordinator.Me(),
	)
	if err != nil {
		return errors.Wrap(err, "authorization failed with error")
	}
	if !isAuthorized {
		return errors.New("can't execute this object")
	}
	return nil
}

func loggerWithTargetID(ctx context.Context, msg insolar.Parcel) context.Context {
	ctx, _ = inslogger.WithField(ctx, "targetid", msg.DefaultTarget().String())
	return ctx
}

// values here (boolean flags) are inverted here, since it's common "predicate" checking function
func noLoopCheckerPredicate(current *Transcript, args interface{}) bool {
	apiReqID := args.(string)
	if current.SentResult ||
		current.Request.ReturnMode == record.ReturnNoWait ||
		current.Request.APIRequestID != apiReqID {
		return true
	}
	return false
}

func (lr *LogicRunner) CheckExecutionLoop(
	ctx context.Context, es *ExecutionState, parcel insolar.Parcel) bool {
	if es.CurrentList.Empty() {
		return false
	}

	msg, ok := parcel.Message().(*message.CallMethod)
	if ok && msg.ReturnMode == record.ReturnNoWait {
		return false
	}

	if es.CurrentList.Check(noLoopCheckerPredicate, msg.APIRequestID) {
		return false
	}

	inslogger.FromContext(ctx).Debug("loop detected")
	return true
}

// finishPendingIfNeeded checks whether last execution was a pending one.
// If this is true as a side effect the function sends a PendingFinished
// message to the current executor
func (lr *LogicRunner) finishPendingIfNeeded(ctx context.Context, es *ExecutionState) {
	es.Lock()
	defer es.Unlock()

	if es.pending != message.InPending {
		return
	}

	es.pending = message.NotPending
	es.PendingConfirmed = false

	pulseObj := lr.pulse(ctx)
	meCurrent, _ := lr.JetCoordinator.IsAuthorized(
		ctx, insolar.DynamicRoleVirtualExecutor, *es.Ref.Record(), pulseObj.PulseNumber, lr.JetCoordinator.Me(),
	)
	if !meCurrent {
		go func() {
			msg := message.PendingFinished{Reference: es.Ref}
			_, err := lr.MessageBus.Send(ctx, &msg, nil)
			if err != nil {
				inslogger.FromContext(ctx).Error("Unable to send PendingFinished message:", err)
			}
		}()
	}
}

func (lr *LogicRunner) executeOrValidate(ctx context.Context, es *ExecutionState, current *Transcript) {

	inslogger.FromContext(ctx).Debug("executeOrValidate")

	ctx, span := instracer.StartSpan(ctx, "LogicRunner.ExecuteOrValidate")
	defer span.End()

	var re insolar.Reply
	var err error
	switch current.Request.CallType {
	case record.CTMethod:
		re, err = lr.executeMethodCall(ctx, es, current)

	case record.CTSaveAsChild, record.CTSaveAsDelegate:
		re, err = lr.executeConstructorCall(ctx, es, current)

	default:
		panic("Unknown e type")
	}
	errstr := ""
	if err != nil {
		inslogger.FromContext(ctx).Warn("contract execution error: ", err)
		errstr = err.Error()
	}

	es.Lock()
	defer es.Unlock()

	current.SentResult = true
	if current.Request.ReturnMode != record.ReturnResult {
		return
	}

	target := *current.RequesterNode
	request := *current.RequestRef
	seq := current.Request.Sequence

	go func() {
		inslogger.FromContext(ctx).Debugf("Sending Method Results for %#v", request)

		_, err := lr.MessageBus.Send(
			ctx,
			&message.ReturnResults{
				Caller:   lr.NodeNetwork.GetOrigin().ID(),
				Target:   target,
				Sequence: seq,
				Reply:    re,
				Error:    errstr,
			},
			&insolar.MessageSendOptions{
				Receiver: &target,
			},
		)
		if err != nil {
			inslogger.FromContext(ctx).Error("couldn't deliver results: ", err)
		}
	}()
}

func (lr *LogicRunner) executeMethodCall(ctx context.Context, es *ExecutionState, current *Transcript) (insolar.Reply, error) {
	ctx, span := instracer.StartSpan(ctx, "LogicRunner.executeMethodCall")
	defer span.End()

	request := current.Request

	objDesc, err := lr.ArtifactManager.GetObject(ctx, *request.Object)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get object")
	}
	current.ObjectDescriptor = objDesc

	if es.PrototypeDescriptor == nil {
		protoDesc, codeDesc, err := lr.DescriptorsCache.ByObjectDescriptor(ctx, objDesc)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get descriptors by prototype reference")
		}

		es.PrototypeDescriptor = protoDesc
		es.CodeDescriptor = codeDesc
	}

	current.LogicContext.Prototype = es.PrototypeDescriptor.HeadRef()
	current.LogicContext.Code = es.CodeDescriptor.Ref()
	current.LogicContext.Parent = current.ObjectDescriptor.Parent()
	// it's needed to assure that we call method on ref, that has same prototype as proxy, that we import in contract code
	if request.Prototype != nil && !request.Prototype.Equal(*es.PrototypeDescriptor.HeadRef()) {
		return nil, errors.New("proxy call error: try to call method of prototype as method of another prototype")
	}

	executor, err := lr.GetExecutor(es.CodeDescriptor.MachineType())
	if err != nil {
		return nil, es.WrapError(current, err, "no executor registered")
	}

	newData, result, err := executor.CallMethod(
		ctx, current.LogicContext, *es.CodeDescriptor.Ref(), current.ObjectDescriptor.Memory(), request.Method, request.Arguments,
	)
	if err != nil {
		return nil, es.WrapError(current, err, "executor error")
	}

	am := lr.ArtifactManager
	if current.Deactivate {
		_, err := am.DeactivateObject(
			ctx, *current.RequestRef, current.ObjectDescriptor, result,
		)
		if err != nil {
			return nil, es.WrapError(current, err, "couldn't deactivate object")
		}
	} else if !bytes.Equal(current.ObjectDescriptor.Memory(), newData) {
		_, err := am.UpdateObject(
			ctx, *current.RequestRef, current.ObjectDescriptor, newData, result,
		)
		if err != nil {
			return nil, es.WrapError(current, err, "couldn't update object")
		}
	} else {
		_, err = am.RegisterResult(ctx, *request.Object, *current.RequestRef, result)
		if err != nil {
			return nil, es.WrapError(current, err, "couldn't save results")
		}
	}

	return &reply.CallMethod{Result: result}, nil
}

func (lr *LogicRunner) executeConstructorCall(
	ctx context.Context, es *ExecutionState, current *Transcript,
) (
	insolar.Reply, error,
) {
	ctx, span := instracer.StartSpan(ctx, "LogicRunner.executeConstructorCall")
	defer span.End()

	request := current.Request

	if current.LogicContext.Caller.IsEmpty() {
		return nil, es.WrapError(current, nil, "Call constructor from nowhere")
	}

	if request.Prototype == nil {
		return nil, es.WrapError(current, nil, "prototype reference is required")
	}

	protoDesc, codeDesc, err := lr.DescriptorsCache.ByPrototypeRef(ctx, *request.Prototype)
	if err != nil {
		return nil, es.WrapError(current, err, "couldn't descriptors")
	}

	current.LogicContext.Prototype = protoDesc.HeadRef()
	current.LogicContext.Code = codeDesc.Ref()

	executor, err := lr.GetExecutor(codeDesc.MachineType())
	if err != nil {
		return nil, es.WrapError(current, err, "no executer registered")
	}

	newData, err := executor.CallConstructor(ctx, current.LogicContext, *codeDesc.Ref(), request.Method, request.Arguments)
	if err != nil {
		return nil, es.WrapError(current, err, "executer error")
	}

	switch request.CallType {
	case record.CTSaveAsChild, record.CTSaveAsDelegate:
		_, err = lr.ArtifactManager.ActivateObject(
			ctx,
			*current.RequestRef, *request.Base, *request.Prototype, request.CallType == record.CTSaveAsDelegate, newData,
		)
		if err != nil {
			return nil, es.WrapError(current, err, "couldn't activate object")
		}
		return &reply.CallConstructor{Object: current.RequestRef}, err

	default:
		return nil, es.WrapError(current, nil, "unsupported type of save object")
	}
}

func (lr *LogicRunner) startGetLedgerPendingRequest(ctx context.Context, es *ExecutionState) {
	err := lr.publisher.Publish(InnerMsgTopic, makeWMMessage(ctx, es.Ref.Bytes(), getLedgerPendingRequestMsg))
	if err != nil {
		inslogger.FromContext(ctx).Warnf("can't send getLedgerPendingRequestMsg: ", err)
	}
}

func (lr *LogicRunner) OnPulse(ctx context.Context, pulse insolar.Pulse) error {
	lr.stateMutex.Lock()

	lr.FlowDispatcher.ChangePulse(ctx, pulse)
	lr.innerFlowDispatcher.ChangePulse(ctx, pulse)

	ctx, span := instracer.StartSpan(ctx, "pulse.logicrunner")
	defer span.End()

	messages := make([]insolar.Message, 0)

	for ref, state := range lr.state {
		meNext, _ := lr.JetCoordinator.IsAuthorized(
			ctx, insolar.DynamicRoleVirtualExecutor, *ref.Record(), pulse.PulseNumber, lr.JetCoordinator.Me(),
		)
		state.Lock()

		if es := state.ExecutionState; es != nil {
			es.Lock()

			toSend := es.OnPulse(ctx, meNext)
			messages = append(messages, toSend...)

			if !meNext {
				if es.CurrentList.Empty() {
					state.ExecutionState = nil
				}
			} else if es.pending == message.NotPending && es.LedgerHasMoreRequests {
				lr.startGetLedgerPendingRequest(ctx, es)
			}

			es.Unlock()
		}

		if state.ExecutionState == nil && state.Validation == nil {
			delete(lr.state, ref)
		}

		state.Unlock()
	}

	lr.stateMutex.Unlock()

	if len(messages) > 0 {
		go lr.sendOnPulseMessagesAsync(ctx, messages)
	}

	lr.stopIfNeeded(ctx)

	return nil
}

func (lr *LogicRunner) stopIfNeeded(ctx context.Context) {
	// lock is required to access lr.state
	lr.stateMutex.Lock()
	defer lr.stateMutex.Unlock()

	if len(lr.state) == 0 {
		lr.stopLock.Lock()
		if lr.isStopping {
			inslogger.FromContext(ctx).Debug("LogicRunner ready to stop")
			lr.stopChan <- struct{}{}
		}
		lr.stopLock.Unlock()
	}
}

func (lr *LogicRunner) sendOnPulseMessagesAsync(ctx context.Context, messages []insolar.Message) {
	ctx, spanMessages := instracer.StartSpan(ctx, "pulse.logicrunner sending messages")
	spanMessages.AddAttributes(trace.StringAttribute("numMessages", strconv.Itoa(len(messages))))

	var sendWg sync.WaitGroup
	sendWg.Add(len(messages))

	for _, msg := range messages {
		go lr.sendOnPulseMessage(ctx, msg, &sendWg)
	}

	sendWg.Wait()
	spanMessages.End()
}

func (lr *LogicRunner) sendOnPulseMessage(ctx context.Context, msg insolar.Message, sendWg *sync.WaitGroup) {
	defer sendWg.Done()
	_, err := lr.MessageBus.Send(ctx, msg, nil)
	if err != nil {
		inslogger.FromContext(ctx).Error(errors.Wrap(err, "error while sending validation data on pulse"))
	}
}

func convertQueueToMessageQueue(ctx context.Context, queue []*Transcript) []message.ExecutionQueueElement {
	mq := make([]message.ExecutionQueueElement, 0)
	var traces string
	for _, elem := range queue {
		mq = append(mq, message.ExecutionQueueElement{
			Parcel:  elem.Parcel,
			Request: elem.RequestRef,
		})

		traces += inslogger.TraceID(elem.Context) + ", "
	}

	inslogger.FromContext(ctx).Debug("convertQueueToMessageQueue: ", traces)

	return mq
}
