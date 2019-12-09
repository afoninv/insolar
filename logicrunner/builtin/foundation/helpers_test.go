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

package foundation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrimPublicKey(t *testing.T) {
	for _, tc := range []struct {
		input  string
		result string
	}{
		{
			input:  "asDafasf",
			result: "asDafasf",
		},
		{
			input:  "-----BEGIN RSA PUBLIC KEY-----\naSDafasf\n-----END RSA PUBLIC KEY-----",
			result: "aSDafasf",
		},
	} {
		require.Equal(t, tc.result, TrimPublicKey(tc.input))
	}
}
