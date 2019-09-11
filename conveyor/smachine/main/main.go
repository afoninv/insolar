///
//    Copyright 2019 Insolar Technologies
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
///

package main

import (
	"context"
	"fmt"
	"github.com/insolar/insolar/conveyor/smachine"
	"github.com/insolar/insolar/conveyor/smachine/main/example"
	"github.com/insolar/insolar/conveyor/smachine/tools"
	"time"
)

func main() {
	var fn1 func(example.ServiceA, string) string
	fn1 = example.ServiceA.DoSomething

	fn1(&implA{}, "test")

	sm := smachine.NewSlotMachine(smachine.SlotMachineConfig{
		SlotPageSize:    10,
		PollingPeriod:   10 * time.Millisecond,
		PollingTruncate: 1 * time.Millisecond,
	}, nil, nil)

	example.SetInjectServiceAdapterA(&implA{}, &sm)

	sm.AddNew(context.Background(), smachine.NoLink(), &example.StateMachine1{})

	signal := tools.NewVersionedSignal()
	worker := smachine.NewSimpleSlotWorker(signal.Mark())

	for i := 0; ; i++ {
		sm.ScanOnce(worker)
		time.Sleep(1 * time.Millisecond)
		if i%100 == 0 {
			fmt.Printf("%03d %v ================================== slots=%v of %v\n", i, time.Now(), sm.OccupiedSlotCount(), sm.AllocatedSlotCount())
			sm.Cleanup()
			fmt.Printf("%03d %v ================================== slots=%v of %v\n", i, time.Now(), sm.OccupiedSlotCount(), sm.AllocatedSlotCount())
		}
	}
}

type implA struct {
}

func (*implA) DoSomething(param string) string {
	return param
}

func (*implA) DoSomethingElse(param0 string, param1 int) (bool, string) {
	return param1 != 0, param0
}