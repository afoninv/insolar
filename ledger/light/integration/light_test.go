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

package integration_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/gen"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
)

func Test_BootstrapCalls(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	cfg := DefaultLightConfig()
	s, err := NewServer(ctx, cfg, nil)
	require.NoError(t, err)

	t.Run("message before pulse received returns error", func(t *testing.T) {
		p, _ := callSetCode(ctx, s)
		_, ok := p.(*payload.Error)
		assert.True(t, ok)
	})

	// First pulse goes in storage then interrupts.
	s.Pulse(ctx)
	// Second pulse goes in storage and starts processing, including pulse change in flow dispatcher.
	s.Pulse(ctx)

	t.Run("messages after two pulses return result", func(t *testing.T) {
		p, _ := callSetCode(ctx, s)
		requireNotError(t, p)
	})
}

func Test_BasicOperations(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	cfg := DefaultLightConfig()
	s, err := NewServer(ctx, cfg, nil)
	require.NoError(t, err)

	// First pulse goes in storage then interrupts.
	s.Pulse(ctx)
	// Second pulse goes in storage and starts processing, including pulse change in flow dispatcher.
	s.Pulse(ctx)

	runner := func(t *testing.T) {
		// Creating root reason request.
		var reasonID insolar.ID
		{
			p := retryIfCancelled(func() payload.Payload {
				p, _ := callSetIncomingRequest(ctx, s, gen.ID(), gen.ID(), true, true)
				return p
			})
			requireNotError(t, p)
			reasonID = p.(*payload.RequestInfo).RequestID
		}
		// Save and check code.
		{
			var sent record.Virtual
			p := retryIfCancelled(func() payload.Payload {
				p, s := callSetCode(ctx, s)
				sent = s
				return p
			})
			requireNotError(t, p)

			p = callGetCode(ctx, s, p.(*payload.ID).ID)
			requireNotError(t, p)
			material := record.Material{}
			err := material.Unmarshal(p.(*payload.Code).Record)
			require.NoError(t, err)
			require.Equal(t, sent, material.Virtual)
		}
		var objectID insolar.ID
		// Set, get request.
		{
			var sent record.Virtual
			p := retryIfCancelled(func() payload.Payload {
				p, s := callSetIncomingRequest(ctx, s, gen.ID(), reasonID, true, true)
				sent = s
				return p
			})
			requireNotError(t, p)

			p = callGetRequest(ctx, s, p.(*payload.RequestInfo).RequestID)
			requireNotError(t, p)
			require.Equal(t, sent, p.(*payload.Request).Request)
			objectID = p.(*payload.Request).RequestID
		}
		// Activate and check object.
		{
			var state record.Virtual
			p := retryIfCancelled(func() payload.Payload {
				p, s := callActivateObject(ctx, s, objectID)
				state = s
				return p
			})
			requireNotError(t, p)
			_, material := requireGetObject(ctx, t, s, objectID)
			require.Equal(t, state, material.Virtual)
		}
		// Amend and check object.
		{
			p := retryIfCancelled(func() payload.Payload {
				p, _ := callSetIncomingRequest(ctx, s, objectID, reasonID, false, true)
				return p
			})
			requireNotError(t, p)

			var state record.Virtual
			p = retryIfCancelled(func() payload.Payload {
				p, s := callAmendObject(ctx, s, objectID, p.(*payload.RequestInfo).RequestID)
				state = s
				return p
			})
			requireNotError(t, p)

			_, material := requireGetObject(ctx, t, s, objectID)
			require.Equal(t, state, material.Virtual)
		}
		// Deactivate and check object.
		{
			p := retryIfCancelled(func() payload.Payload {
				p, _ := callSetIncomingRequest(ctx, s, objectID, reasonID, false, true)
				return p
			})
			requireNotError(t, p)

			retryIfCancelled(func() payload.Payload {
				p, _ := callDeactivateObject(ctx, s, objectID, p.(*payload.RequestInfo).RequestID)
				return p
			})

			lifeline, _ := callGetObject(ctx, s, objectID)
			_, ok := lifeline.(*payload.Error)
			assert.True(t, ok)
		}
	}

	t.Run("happy basic", runner)

	t.Run("happy concurrent", func(t *testing.T) {
		count := 100
		pulseAt := rand.Intn(count)
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			if i == pulseAt {
				s.Pulse(ctx)
			}
			i := i
			go func() {
				t.Run(fmt.Sprintf("iter %d", i), runner)
				wg.Done()
			}()
		}

		wg.Wait()
	})
}

func requireNotError(t *testing.T, pl payload.Payload) {
	if err, ok := pl.(*payload.Error); ok {
		t.Fatal(err)
	}
}

func requireGetObject(ctx context.Context, t *testing.T, s *Server, objectID insolar.ID) (record.Lifeline, record.Material) {
	lifelinePL, statePL := callGetObject(ctx, s, objectID)
	requireNotError(t, lifelinePL)
	requireNotError(t, statePL)

	lifeline := record.Lifeline{}
	err := lifeline.Unmarshal(lifelinePL.(*payload.Index).Index)
	require.NoError(t, err)

	state := record.Material{}
	err = state.Unmarshal(statePL.(*payload.State).Record)
	require.NoError(t, err)

	return lifeline, state
}

func retryIfCancelled(cb func() payload.Payload) payload.Payload {
	rep := cb()
	if err, ok := rep.(*payload.Error); ok {
		if err.Code == payload.CodeFlowCanceled {
			return retryIfCancelled(cb)
		}
	}

	return rep
}
