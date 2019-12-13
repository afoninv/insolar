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

package sm_object

import (
	"fmt"

	"github.com/insolar/insolar/conveyor/smachine"
	"github.com/insolar/insolar/insolar"
)

type ObjectPair struct {
	Pulse           insolar.PulseNumber
	ObjectReference insolar.Reference
}

func (p ObjectPair) String() string {
	return fmt.Sprintf("object-%d-%s", p.Pulse, p.ObjectReference)
}

type LocalObjectCatalog struct{}

func (p LocalObjectCatalog) Get(ctx smachine.ExecutionContext, pair ObjectPair) SharedObjectStateAccessor {
	if v, ok := p.TryGet(ctx, pair); ok {
		return v
	}
	panic(fmt.Sprintf("missing entry: %s", pair.String()))
}

func (p LocalObjectCatalog) TryGet(ctx smachine.ExecutionContext, pair ObjectPair) (SharedObjectStateAccessor, bool) {
	if v := ctx.GetPublishedLink(pair); v.IsAssignableTo((*SharedObjectState)(nil)) {
		return SharedObjectStateAccessor{v}, true
	}
	return SharedObjectStateAccessor{}, false
}

func (p LocalObjectCatalog) GetOrCreate(ctx smachine.ExecutionContext, pair ObjectPair) SharedObjectStateAccessor {
	if v, ok := p.TryGet(ctx, pair); ok {
		return v
	}

	ctx.InitChild(func(ctx smachine.ConstructionContext) smachine.StateMachine {
		return NewObjectSM(pair.ObjectReference)
	})

	return p.Get(ctx, pair)
}

// //////////////////////////////////////

type SharedObjectStateAccessor struct {
	smachine.SharedDataLink
}

func (v SharedObjectStateAccessor) Prepare(fn func(*SharedObjectState)) smachine.SharedDataAccessor {
	return v.PrepareAccess(func(data interface{}) bool {
		fn(data.(*SharedObjectState))
		return false
	})
}