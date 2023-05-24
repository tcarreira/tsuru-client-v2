// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFakeExecCommand(t *testing.T) {
	t.Run("empty fake exec", func(t *testing.T) {
		fakeE := FakeExec{}
		err := fakeE.Command(ExecuteOptions{})
		assert.NoError(t, err)
	})

	t.Run("fake exec with output", func(t *testing.T) {
		fakeE := FakeExec{
			outStderr: "error output",
			outStdout: "standard output",
			outErr:    fmt.Errorf("error"),
		}
		stderr, stdout := bytes.Buffer{}, bytes.Buffer{}
		err := fakeE.Command(ExecuteOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		})
		assert.ErrorIs(t, err, fakeE.outErr)
		assert.Equal(t, fakeE.outStdout, stdout.String())
		assert.Equal(t, fakeE.outStderr, stderr.String())
	})
}
