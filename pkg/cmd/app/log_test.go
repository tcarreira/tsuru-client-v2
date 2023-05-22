// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppLogIsRegistered(t *testing.T) {
	appCmd := NewAppCmd()
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "log" {
			found = true
			break
		}
	}
	assert.True(t, found, "subcommand log not registered in appCmd")
}
