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

package sm_execute_request

import (
	"github.com/pkg/errors"

	"github.com/insolar/insolar/conveyor"
	"github.com/insolar/insolar/conveyor/injector"
	"github.com/insolar/insolar/conveyor/smachine"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/logicrunner/artifacts"
	"github.com/insolar/insolar/logicrunner/common"
	"github.com/insolar/insolar/logicrunner/s_artifact"
	"github.com/insolar/insolar/logicrunner/s_contract_runner"
	"github.com/insolar/insolar/logicrunner/s_sender"
	"github.com/insolar/insolar/logicrunner/sm_object"
)

type ExecuteIncomingCommon struct {
	objectCatalog  *sm_object.LocalObjectCatalog
	pulseSlot      *conveyor.PulseSlot
	ArtifactClient *s_artifact.ArtifactClientServiceAdapter
	Sender         *s_sender.SenderServiceAdapter
	ContractRunner *s_contract_runner.ContractRunnerServiceAdapter

	objectInfo sm_object.ObjectInfo

	sharedStateLink sm_object.SharedObjectStateAccessor

	internalError error

	// input
	MessageMeta *payload.Meta
	Request     *record.IncomingRequest

	// values to pass between steps
	RequestReference       insolar.Reference
	RequestObjectReference insolar.Reference
	RequestDeduplicated    bool
	DeduplicatedResult     *record.Result

	contractTranscript  *common.Transcript
	executionResult     artifacts.RequestResult
	newObjectDescriptor artifacts.ObjectDescriptor
}

func (s *ExecuteIncomingCommon) InjectDependencies(sm smachine.StateMachine, slotLink smachine.SlotLink, injector *injector.DependencyInjector) {
	injector.MustInject(&s.ArtifactClient)
	injector.MustInject(&s.Sender)
	injector.MustInject(&s.ContractRunner)

	injector.MustInject(&s.pulseSlot)
	injector.MustInject(&s.objectCatalog)
}

func (s *ExecuteIncomingCommon) useSharedObjectInfo(ctx smachine.ExecutionContext, cb func(state *sm_object.SharedObjectState)) smachine.StateUpdate {
	goCtx := ctx.GetContext()

	if s.sharedStateLink.IsZero() {
		objectPair := sm_object.ObjectPair{
			Pulse:           s.pulseSlot.PulseData().PulseNumber,
			ObjectReference: s.RequestObjectReference,
		}
		s.sharedStateLink = s.objectCatalog.GetOrCreate(ctx, objectPair)
	}

	switch s.sharedStateLink.Prepare(cb).TryUse(ctx).GetDecision() {
	case smachine.NotPassed:
		inslogger.FromContext(goCtx).Error("NotPassed")
		return ctx.WaitShared(s.sharedStateLink.SharedDataLink).ThenRepeat()
	case smachine.Impossible:
		inslogger.FromContext(goCtx).Error("Impossible")
		// the holder of the sharedState is stopped
		return ctx.Stop()
	case smachine.Passed:
		inslogger.FromContext(goCtx).Error("Passed")
	default:
		panic("unknown state from TryUse")
	}
	return smachine.StateUpdate{}
}

func (s *ExecuteIncomingCommon) internalStepSaveResult(ctx smachine.ExecutionContext, fetchNew bool) smachine.ConditionalBuilder {
	goCtx := ctx.GetContext()

	objectReference := s.RequestObjectReference
	requestReference := s.RequestReference
	executionResult := s.executionResult

	return s.ArtifactClient.PrepareAsync(ctx, func(svc s_artifact.ArtifactClientService) smachine.AsyncResultFunc {
		var objectDescriptor artifacts.ObjectDescriptor

		err := svc.RegisterResult(goCtx, requestReference, executionResult)
		if err == nil && fetchNew {
			objectDescriptor, err = svc.GetObject(goCtx, objectReference, nil)
		}

		return func(ctx smachine.AsyncResultContext) {
			s.internalError = err
			if objectDescriptor != nil {
				s.newObjectDescriptor = objectDescriptor
			}
		}
	}).WithFlags(smachine.AutoWakeUp).DelayedStart().Sleep()
}

// it'll panic or execute
func (s *ExecuteIncomingCommon) internalSendResult(ctx smachine.ExecutionContext) {
	goCtx := ctx.GetContext()

	var executionBytes []byte
	var executionError string

	switch {
	case s.internalError != nil: // execution error
		executionError = s.internalError.Error()
	case s.DeduplicatedResult != nil: // result of deduplication
		if s.executionResult != nil {
			panic("we got deduplicated result and execution result, unreachable")
		}
		material := record.Material{}
		if err := material.Unmarshal(s.DeduplicatedResult.Payload); err != nil {
			executionError = errors.Wrap(err, "failed to unmarshal deduplicated result").Error()
			break
		}

		virtual := record.Unwrap(&material.Virtual)
		result, ok := virtual.(*record.Result)
		if !ok {
			executionError = errors.Errorf("unexpected record %T", virtual).Error()
			break
		}

		executionBytes = result.Payload
	case s.executionResult != nil: // result of execution
		executionBytes = s.executionResult.Result()
	default:
		// we have no result and no error (??)
		panic("unreachable")
	}

	pl := &payload.ReturnResults{
		RequestRef: s.RequestReference,
		Reply:      executionBytes,
		Error:      executionError,
	}

	APIRequest := s.Request.APINode.IsEmpty()
	if !APIRequest {
		pl.Target = s.Request.Caller
		pl.Reason = s.Request.Reason
	}

	msg, err := payload.NewResultMessage(pl)
	if err != nil {
		panic("couldn't serialize message: " + err.Error())
	}

	request := s.Request
	s.Sender.PrepareNotify(ctx, func(svc s_sender.SenderService) {
		// TODO[bigbes]: there should be retry sender
		// retrySender := bus.NewWaitOKWithRetrySender(svc, svc, 1)

		var done func()
		if APIRequest {
			_, done = svc.SendTarget(goCtx, msg, request.APINode)
		} else {
			_, done = svc.SendRole(goCtx, msg, insolar.DynamicRoleVirtualExecutor, request.Caller)
		}
		done()
	})
}

func (s *ExecuteIncomingCommon) stepStop(ctx smachine.ExecutionContext) smachine.StateUpdate {
	if s.internalError != nil {
		return ctx.Jump(s.stepError)
	}
	return ctx.Stop()
}

func (s *ExecuteIncomingCommon) stepError(ctx smachine.ExecutionContext) smachine.StateUpdate {
	return ctx.Error(s.internalError)
}