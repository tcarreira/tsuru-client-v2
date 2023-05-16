// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppShellInfo(t *testing.T) {
	var stdout bytes.Buffer
	appShellCmd := newAppShellCmd()
	appShellCmd.SetOutput(&stdout)
	err := appShellCmd.Help()
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
}

func TestAppShellIsRegistered(t *testing.T) {
	appCmd := NewAppCmd()
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "shell" {
			found = true
			break
		}
	}
	assert.True(t, found, "subcommand list not registered in appCmd")
}

