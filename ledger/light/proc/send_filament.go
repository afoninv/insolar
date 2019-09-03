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

	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/instrumentation/instracer"
	"github.com/insolar/insolar/ledger/light/executor"
)

type SendFilament struct {
	message          payload.Meta
	objID, startFrom insolar.ID
	readUntil        insolar.PulseNumber

	dep struct {
		sender    bus.Sender
		filaments executor.FilamentCalculator
	}
}

func NewSendFilament(msg payload.Meta, objID insolar.ID, startFrom insolar.ID, readUntil insolar.PulseNumber) *SendFilament {
	return &SendFilament{
		message:   msg,
		objID:     objID,
		startFrom: startFrom,
		readUntil: readUntil,
	}
}

func (p *SendFilament) Dep(sender bus.Sender, filaments executor.FilamentCalculator) {
	p.dep.sender = sender
	p.dep.filaments = filaments
}

func (p *SendFilament) Proceed(ctx context.Context) error {
	ctx, span := instracer.StartSpan(ctx, "SendFilament")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("objID", p.objID.DebugString()),
		trace.StringAttribute("startFrom", p.startFrom.DebugString()),
		trace.StringAttribute("readUntil", p.readUntil.String()),
	)

	records, err := p.dep.filaments.Requests(ctx, p.objID, p.startFrom, p.readUntil)
	if err != nil {
		return errors.Wrap(err, "failed to fetch filament")
	}
	if len(records) == 0 {
		return p.replyError(ctx, p.message, "requests not found", payload.CodeNotFound)
	}

	msg, err := payload.NewMessage(&payload.FilamentSegment{
		ObjectID: p.objID,
		Records:  records,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create message")
	}
	p.dep.sender.Reply(ctx, p.message, msg)
	return nil
}

func (p *SendFilament) replyError(
	ctx context.Context,
	inputMessage payload.Meta,
	text string,
	code uint32,
) error {
	msg, err := payload.NewMessage(&payload.Error{
		Text: text,
		Code: code,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create reply")
	}

	p.dep.sender.Reply(ctx, inputMessage, msg)
	return nil
}
