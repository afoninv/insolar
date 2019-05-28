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

package proc

import (
	"context"
	"fmt"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/delegationtoken"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/flow/bus"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/blob"
	"github.com/insolar/insolar/ledger/object"
	"github.com/insolar/insolar/messagebus"
	"github.com/pkg/errors"
)

type SendObject struct {
	message        bus.Message
	jet            insolar.JetID
	index          object.Lifeline
	returnPendings int

	Dep struct {
		Coordinator    jet.Coordinator
		Jets           jet.Storage
		JetUpdater     jet.Fetcher
		RecordAccessor object.RecordAccessor
		Blobs          blob.Storage
		Bus            insolar.MessageBus

		PendingAccessor object.PendingAccessor
		PendingModifier object.PendingModifier
	}
}

func NewSendObject(msg bus.Message, jet insolar.JetID, idx object.Lifeline) *SendObject {
	return &SendObject{
		message: msg,
		jet:     jet,
		index:   idx,
	}
}

func (p *SendObject) Proceed(ctx context.Context) error {
	r := bus.Reply{}
	r.Reply, r.Err = p.handle(ctx, p.message.Parcel)
	p.message.ReplyTo <- r
	return r.Err
}

func (p *SendObject) handle(
	ctx context.Context, parcel insolar.Parcel,
) (insolar.Reply, error) {
	msg := parcel.Message().(*message.GetObject)
	logger := inslogger.FromContext(ctx)

	var stateID *insolar.ID
	if msg.State != nil {
		stateID = msg.State
	} else {
		stateID = p.index.LatestState
	}
	if stateID == nil {
		return &reply.Error{ErrType: reply.ErrStateNotAvailable}, nil
	}

	var (
		stateJet *insolar.ID
	)
	onHeavy, err := p.Dep.Coordinator.IsBeyondLimit(ctx, parcel.Pulse(), stateID.Pulse())
	if err != nil && err != pulse.ErrNotFound {
		return nil, err
	}
	if onHeavy {
		hNode, err := p.Dep.Coordinator.Heavy(ctx, parcel.Pulse())
		if err != nil {
			return nil, err
		}
		logger.WithFields(map[string]interface{}{
			"state":    stateID.DebugString(),
			"going_to": hNode.String(),
		}).Debug("fetching object (on heavy)")

		obj, err := p.fetchObject(ctx, msg.Head, *hNode, stateID, parcel.Pulse())
		if err != nil {
			if err == insolar.ErrDeactivated {
				return &reply.Error{ErrType: reply.ErrDeactivated}, nil
			}
			return nil, err
		}

		return &reply.Object{
			Head:         msg.Head,
			State:        *stateID,
			Prototype:    obj.Prototype,
			IsPrototype:  obj.IsPrototype,
			ChildPointer: p.index.ChildPointer,
			Parent:       p.index.Parent,
			Memory:       obj.Memory,
			Pendings:     []record.Request{},
		}, nil
	}

	stateJetID, actual := p.Dep.Jets.ForID(ctx, stateID.Pulse(), *msg.Head.Record())
	stateJet = (*insolar.ID)(&stateJetID)

	if !actual {
		actualJet, err := p.Dep.JetUpdater.Fetch(ctx, *msg.Head.Record(), stateID.Pulse())
		if err != nil {
			return nil, err
		}
		stateJet = actualJet
	}

	// Fetch pendings
	pendings, err := p.fetchPendings(ctx, parcel.Pulse(), *msg.Head.Record())

	// Fetch state record.
	rec, err := p.Dep.RecordAccessor.ForID(ctx, *stateID)

	if err == object.ErrNotFound {
		// The record wasn't found on the current suitNode. Return redirect to the node that contains it.
		// We get Jet tree for pulse when given state was added.
		suitNode, err := p.Dep.Coordinator.NodeForJet(ctx, *stateJet, parcel.Pulse(), stateID.Pulse())
		if err != nil {
			return nil, err
		}
		logger.WithFields(map[string]interface{}{
			"state":    stateID.DebugString(),
			"going_to": suitNode.String(),
		}).Debug("fetching object (record not found)")

		obj, err := p.fetchObject(ctx, msg.Head, *suitNode, stateID, parcel.Pulse())
		if err != nil {
			if err == insolar.ErrDeactivated {
				return &reply.Error{ErrType: reply.ErrDeactivated}, nil
			}
			return nil, err
		}

		return &reply.Object{
			Head:         msg.Head,
			State:        *stateID,
			Prototype:    obj.Prototype,
			IsPrototype:  obj.IsPrototype,
			ChildPointer: p.index.ChildPointer,
			Parent:       p.index.Parent,
			Memory:       obj.Memory,
			Pendings:     pendings,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "can't fetch record from storage")
	}

	virtRec := rec.Virtual
	concrete := record.Unwrap(virtRec)
	state, ok := concrete.(record.State)
	if !ok {
		return nil, fmt.Errorf("invalid object record %#v", virtRec)
	}

	if state.ID() == record.StateDeactivation {
		return &reply.Error{ErrType: reply.ErrDeactivated}, nil
	}

	var childPointer *insolar.ID
	if p.index.ChildPointer != nil {
		childPointer = p.index.ChildPointer
	}
	rep := reply.Object{
		Head:         msg.Head,
		State:        *stateID,
		Prototype:    state.GetImage(),
		IsPrototype:  state.GetIsPrototype(),
		ChildPointer: childPointer,
		Parent:       p.index.Parent,
		Pendings:     pendings,
	}

	if state.GetMemory() != nil && state.GetMemory().NotEmpty() {
		b, err := p.Dep.Blobs.ForID(ctx, *state.GetMemory())
		if err == blob.ErrNotFound {
			hNode, err := p.Dep.Coordinator.Heavy(ctx, parcel.Pulse())
			if err != nil {
				return nil, err
			}
			obj, err := p.fetchObject(ctx, msg.Head, *hNode, stateID, parcel.Pulse())
			if err != nil {
				return nil, err
			}
			err = p.Dep.Blobs.Set(ctx, *state.GetMemory(), blob.Blob{
				JetID: p.jet,
				Value: obj.Memory},
			)
			if err != nil {
				return nil, err
			}
			b.Value = obj.Memory
		}
		rep.Memory = b.Value
	}

	return &rep, nil
}

func (p *SendObject) fetchObject(
	ctx context.Context, obj insolar.Reference, node insolar.Reference, stateID *insolar.ID, pulse insolar.PulseNumber,
) (*reply.Object, error) {
	sender := messagebus.BuildSender(
		p.Dep.Bus.Send,
		messagebus.FollowRedirectSender(p.Dep.Bus),
		messagebus.RetryJetSender(p.Dep.Jets),
	)
	genericReply, err := sender(
		ctx,
		&message.GetObject{
			Head:  obj,
			State: stateID,
		},
		&insolar.MessageSendOptions{
			Receiver: &node,
			Token:    &delegationtoken.GetObjectRedirectToken{},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch object state")
	}
	if rep, ok := genericReply.(*reply.Error); ok {
		return nil, rep.Error()
	}

	rep, ok := genericReply.(*reply.Object)
	if !ok {
		return nil, fmt.Errorf("failed to fetch object state: unexpected reply type %T (reply=%+v)", genericReply, genericReply)
	}
	return rep, nil
}

func (p *SendObject) fetchPendings(ctx context.Context, currentPN insolar.PulseNumber, objID insolar.ID) ([]record.Request, error) {
	pendMeta, err := p.Dep.PendingAccessor.MetaForObjID(ctx, currentPN, objID)
	if err != nil {
		return []record.Request{}, err
	}

	// Because we are the first light in the chain
	if pendMeta.PreviousPN == nil || pendMeta.IsChainCompleted == true {
		return p.Dep.PendingAccessor.ForObjID(ctx, currentPN, objID, p.returnPendings)
	}

	if pendMeta.ReadUntil == nil {
		panic("inconsistent state of the pending filament")
	}

	err = p.fillPendingFilament(ctx, currentPN, objID, *pendMeta.PreviousPN, *pendMeta.ReadUntil)
	if err != nil {
		return []record.Request{}, err
	}

	err = p.Dep.PendingModifier.RefreshState(ctx, currentPN, objID)
	if err != nil {
		return []record.Request{}, err
	}

	return p.Dep.PendingAccessor.ForObjID(ctx, currentPN, objID, p.returnPendings)
}

func (p *SendObject) fillPendingFilament(ctx context.Context, currentPN insolar.PulseNumber, objID insolar.ID, destPN insolar.PulseNumber, readUntil insolar.PulseNumber) error {
	continueFilling := true

	for continueFilling {
		node, err := p.Dep.Coordinator.NodeForObject(ctx, objID, currentPN, destPN)
		if err != nil {
			return err
		}

		rep, err := p.Dep.Bus.Send(
			ctx,
			&message.GetPendingFilament{ObjectID: objID},
			&insolar.MessageSendOptions{
				Receiver: node,
			},
		)
		if err != nil {
			return err
		}

		switch r := rep.(type) {
		case *reply.PendingFilament:
			err := p.Dep.PendingModifier.SetFilament(ctx, flow.Pulse(ctx), objID, destPN, r.Records)
			if err != nil {
				return err
			}

			if r.HasFullChain == true {
				continueFilling = false
			}
			if r.PreviousPendingPN == nil {
				continueFilling = false
			} else if *r.PreviousPendingPN > readUntil {
				destPN = *r.PreviousPendingPN
			}
		case *reply.Error:
			return r.Error()
		default:
			return fmt.Errorf("fillPendingFilament: unexpected reply: %#v", rep)
		}
	}

	return nil
}
