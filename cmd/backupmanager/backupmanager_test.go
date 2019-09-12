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
// +build slowtest

package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var binaryPath string

func init() {
	var ok bool

	binaryPath, ok = os.LookupEnv("BIN_DIR")
	if !ok {
		wd, err := os.Getwd()
		binaryPath = filepath.Join(wd, "..", "..", "bin")

		if err != nil {
			panic(err.Error())
		}
	}
}

func invoke(args ...string) (string, error) {
	cmd := exec.Command(binaryPath+"/backupmanager", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestNoMergeToEmptyDb(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	for i := 0; i < 3; i++ {
		output, err := invoke("merge", "-t", tmpdir, "-n", "TEST")
		_, ok := err.(*exec.ExitError)
		require.True(t, ok)
		require.Contains(t, output, "ERROR : db must not be empty")
	}
}

func TestNoCreateToExistingDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	_, err = invoke("create", "-d", tmpdir)
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		output, err := invoke("create", "-d", tmpdir)
		_, ok := err.(*exec.ExitError)
		require.True(t, ok)
		require.Contains(t, output, "ERROR : DB must be empty")
	}
}

func TestNoPrepareBackupToEmptyDb(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	for i := 0; i < 3; i++ {
		output, err := invoke("prepare_backup", "-d", tmpdir, "-l", "TEST")
		_, ok := err.(*exec.ExitError)
		require.True(t, ok)
		require.Contains(t, output, "no backup start keys")
	}
}

func TestFailToMergeBadBackupFile(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	_, err = invoke("create", "-d", tmpdir)
	require.NoError(t, err)

	bkpFile := tmpdir + "/incr.bkp"
	err = ioutil.WriteFile(bkpFile, []byte("test Data"), 0600)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err := invoke("merge", "-t", tmpdir, "-n", bkpFile)
		_, ok := err.(*exec.ExitError)
		require.True(t, ok)
	}
}