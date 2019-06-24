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

// +build functest

package functest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMember(t *testing.T) {
	member, err := newUserWithKeys()
	require.NoError(t, err)
	member.ref = root.ref
	addBurnAddresses(t)
	result, err := signedRequest(member, "contract.createMember", map[string]interface{}{})
	require.NoError(t, err)
	ref, ok := result.(string)
	require.True(t, ok)
	require.NotEqual(t, "", ref)
}

func TestCreateMemberWithoutBurnAddresses(t *testing.T) {
	member, err := newUserWithKeys()
	require.NoError(t, err)
	member.ref = root.ref
	member.pubKey = "fake"
	_, err = signedRequest(member, "contract.createMember", map[string]interface{}{})
	require.Nil(t, err)
}

func TestCreateMemberWithBadKey(t *testing.T) {
	member, err := newUserWithKeys()
	require.NoError(t, err)
	member.ref = root.ref
	member.pubKey = "fake"
	addBurnAddresses(t)
	_, err = signedRequest(member, "contract.createMember", map[string]interface{}{})
	require.Nil(t, err)
}

func TestCreateMemberWhenNoBurnAddressesLeft(t *testing.T) {
	member1, err := newUserWithKeys()
	require.NoError(t, err)
	member1.ref = root.ref
	addBurnAddresses(t)
	_, err = signedRequest(member1, "contract.createMember", map[string]interface{}{})

	member2, err := newUserWithKeys()
	require.NoError(t, err)
	member2.ref = root.ref

	_, err = signedRequest(member2, "contract.createMember", map[string]interface{}{})
	require.Contains(t, err.Error(), "no more burn address left")
}
