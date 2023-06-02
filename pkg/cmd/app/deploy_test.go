// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func TestAppDeployInfo(t *testing.T) {
	stdout := strings.Builder{}
	appDeployCmd := newAppDeployCmd(tsuructx.TsuruContextWithConfig(nil))
	appDeployCmd.SetOutput(&stdout)
	err := appDeployCmd.Help()
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())
}

func TestAppDeployIsRegistered(t *testing.T) {
	appCmd := NewAppCmd(tsuructx.TsuruContextWithConfig(nil))
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "deploy" {
			found = true
			break
		}
	}
	assert.True(t, found, "subcommand deploy not registered in appCmd")
}
