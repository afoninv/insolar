/*
 *    Copyright 2018 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package pulsemanager

import (
	"context"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/core/message"
	"github.com/insolar/insolar/ledger/jetdrop"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/insolar/insolar/log"
)

// PulseManager implements core.PulseManager.
type PulseManager struct {
	db  *storage.DB
	lr  core.LogicRunner
	bus core.MessageBus
}

// Current returns current pulse structure.
func (m *PulseManager) Current() (*core.Pulse, error) {
	latestPulse, err := m.db.GetLatestPulseNumber()
	if err != nil {
		return nil, err
	}
	pulse, err := m.db.GetPulse(latestPulse)
	if err != nil {
		return nil, err
	}
	data := core.Pulse{
		PulseNumber:     latestPulse,
		Entropy:         pulse.Entropy,
		NextPulseNumber: pulse.PredictedNextPulse,
	}
	return &data, nil
}

// Set set's new pulse and closes current jet drop.
func (m *PulseManager) Set(pulse core.Pulse) error {
	log.Infof("pn - %v", pulse.PulseNumber)
	latestPulseNumber, err := m.db.GetLatestPulseNumber()
	if err != nil {
		log.Error("GetLatestPulseNumber")
		log.Error(err)
		return err
	}
	latestPulse, err := m.db.GetPulse(latestPulseNumber)
	if err != nil {
		log.Error("GetPulse")
		log.Error(err)
		return err
	}
	prevDrop, err := m.db.GetDrop(latestPulse.PrevPulse)
	if err != nil {
		log.Error("GetDrop")
		log.Error(err)
		return err
	}
	log.Infof("before create drop pr ne pul - %v", latestPulse.PredictedNextPulse)
	log.Infof("before create drop prev pul - %v", latestPulse.PrevPulse)
	log.Infof("latest pulse - %v", latestPulseNumber)
	drop, messages, err := m.db.CreateDrop(latestPulseNumber, prevDrop.Hash)
	if err != nil {
		log.Error("CreateDrop")
		log.Error(err)
		return err
	}
	err = m.db.SetDrop(drop)
	if err != nil {
		log.Error("SetDrop")
		log.Error(err)
		return err
	}

	dropSerialized, err := jetdrop.Encode(drop)
	if err != nil {
		log.Error("Encode")
		log.Error(err)
		return err
	}

	msg := &message.JetDrop{
		Drop:        dropSerialized,
		Messages:    messages,
		PulseNumber: latestPulseNumber,
	}
	_, err = m.bus.Send(context.TODO(), msg)
	if err != nil {
		log.Error("Send")
		log.Error(err)
		return err
	}
	log.Info("Все хорошо")
	err = m.db.AddPulse(pulse)
	if err != nil {
		return err
	}

	return m.lr.OnPulse(pulse)
}

// NewPulseManager creates PulseManager instance.
func NewPulseManager(db *storage.DB) (*PulseManager, error) {
	pm := PulseManager{db: db}
	return &pm, nil
}

// Link links external components.
func (m *PulseManager) Link(components core.Components) error {
	m.bus = components.MessageBus
	m.lr = components.LogicRunner
	return nil
}
