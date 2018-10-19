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

package member

import (
	"errors"

	"github.com/insolar/insolar/application/contract/member/signer"
	"github.com/insolar/insolar/application/proxy/rootdomain"
	"github.com/insolar/insolar/application/proxy/wallet"
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/cryptohelpers/ecdsa"
	"github.com/insolar/insolar/logicrunner/goplugin/foundation"
)

type Member struct {
	foundation.BaseContract
	Name      string
	PublicKey string
}

func (m *Member) GetName() (string, error) {
	return m.Name, nil
}
func (m *Member) GetPublicKey() (string, error) {
	return m.PublicKey, nil
}

func New(name string, key string) (*Member, error) {
	return &Member{
		Name:      name,
		PublicKey: key,
	}, nil
}

func (m *Member) verifySig(method string, params []byte, seed []byte, sign []byte) error {
	args, err := core.MarshalArgs(
		m.GetReference(),
		method,
		params,
		seed)
	if err != nil {
		return err
	}
	key, err := m.GetPublicKey()
	if err != nil {
		return err
	}
	verified, err := ecdsa.Verify(args, sign, key)
	if err != nil {
		return err
	}
	if !verified {
		return errors.New("Incorrect signature")
	}
	return nil
}

// Call method for authorized calls
func (m *Member) Call(rootDomain core.RecordRef, method string, params []byte, seed []byte, sign []byte) (interface{}, error) {

	if err := m.verifySig(method, params, seed, sign); err != nil {
		return nil, err
	}

	switch method {
	case "CreateMember":
		return m.createMemberCall(rootDomain, params)
	case "GetMyBalance":
		return m.getMyBalance()
	case "GetBalance":
		return m.getBalance(params)
	case "Transfer":
		return m.transferCall(params)
	case "DumpUserInfo":
		return m.dumpUserInfoCall(rootDomain, params)
	case "DumpAllUsers":
		return m.dumpAllUsersCall(rootDomain)
	}
	return nil, &foundation.Error{S: "Unknown method"}
}

func (m *Member) createMemberCall(ref core.RecordRef, params []byte) (interface{}, error) {
	rootDomain := rootdomain.GetObject(ref)
	var name string
	var key string
	if err := signer.UnmarshalParams(params, &name, &key); err != nil {
		return nil, err
	}
	return rootDomain.CreateMember(name, key)
}

func (m *Member) getMyBalance() (interface{}, error) {
	return wallet.GetImplementationFrom(m.GetReference()).GetTotalBalance()
}

func (m *Member) getBalance(params []byte) (interface{}, error) {
	var member string
	if err := signer.UnmarshalParams(params, &member); err != nil {
		return nil, err
	}
	return wallet.GetImplementationFrom(core.NewRefFromBase58(member)).GetTotalBalance()
}

func (m *Member) transferCall(params []byte) (interface{}, error) {
	var amount float64
	var toStr string
	if err := signer.UnmarshalParams(params, &amount, &toStr); err != nil {
		return nil, err
	}
	to := core.NewRefFromBase58(toStr)
	return nil, wallet.GetImplementationFrom(m.GetReference()).Transfer(uint(amount), &to)
}

func (m *Member) dumpUserInfoCall(ref core.RecordRef, params []byte) (interface{}, error) {
	rootDomain := rootdomain.GetObject(ref)
	var user string
	if err := signer.UnmarshalParams(params, &user); err != nil {
		return nil, err
	}
	return rootDomain.DumpUserInfo(user)
}

func (m *Member) dumpAllUsersCall(ref core.RecordRef) (interface{}, error) {
	rootDomain := rootdomain.GetObject(ref)
	return rootDomain.DumpAllUsers()
}
