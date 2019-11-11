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

package sm_request

import (
	"context"
	"fmt"

	"github.com/insolar/insolar/conveyor/smachine"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus/meta"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/logicrunner/common"
)

type logProcessing struct {
	*insolar.LogObjectTemplate `txt:"processing message"`

	message_type string
}

func HandlerFactoryMeta(message *common.DispatcherMessage) smachine.CreateFunc {
	payloadMeta := message.PayloadMeta
	messageMeta := message.MessageMeta
	traceID := messageMeta.Get(meta.TraceID)

	payloadBytes := payloadMeta.Payload
	payloadType, err := payload.UnmarshalType(payloadBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal payload type: %s", err.Error()))
	}

	goCtx, _ := inslogger.WithTraceField(context.Background(), traceID)
	goCtx, logger := inslogger.WithField(goCtx, "component", "sm")

	logger.Error(logProcessing{message_type: payloadType.String()})

	switch payloadType {
	case payload.TypeCallMethod:
		pl := payload.CallMethod{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.CallMethod: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachineCallMethod{Meta: payloadMeta, Payload: &pl}
		}

	case payload.TypeSagaCallAcceptNotification:
		pl := payload.SagaCallAcceptNotification{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.SagaCallAcceptNotification: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachineSagaAccept{Meta: payloadMeta, Payload: &pl}
		}

	case payload.TypeUpdateJet:
		pl := payload.UpdateJet{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.UpdateJet: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachineUpdateJet{Meta: payloadMeta, Payload: &pl}
		}

	case payload.TypePendingFinished:
		pl := payload.PendingFinished{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.PendingFinished: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachinePendingFinished{Meta: payloadMeta, Payload: &pl}
		}

	case payload.TypeExecutorResults:
		pl := payload.ExecutorResults{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.ExecutorResults: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachineExecutorResults{Meta: payloadMeta, Payload: &pl}
		}

	case payload.TypeStillExecuting:
		pl := payload.StillExecuting{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.StillExecuting: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachineStillExecuting{Meta: payloadMeta, Payload: &pl}
		}

	case payload.TypeAdditionalCallFromPreviousExecutor:
		pl := payload.AdditionalCallFromPreviousExecutor{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.AdditionalCallFromPreviousExecutor: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachineAdditionalCall{Meta: payloadMeta, Payload: &pl}
		}

	case payload.TypeAbandonedRequestsNotification:
		pl := payload.AbandonedRequestsNotification{}
		if err := pl.Unmarshal(payloadBytes); err != nil {
			panic(fmt.Sprintf("failed to unmarshal payload.AbandonedRequestsNotification: %s", err.Error()))
		}
		return func(ctx smachine.ConstructionContext) smachine.StateMachine {
			ctx.SetContext(goCtx)
			ctx.SetTracerId(traceID)
			return &StateMachineAbandonedRequests{Meta: payloadMeta, Payload: &pl}
		}

	default:
		panic(fmt.Sprintf(" no handler for message type %s", payloadType.String()))
	}
}
