package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppListIsRegistered(t *testing.T) {
	appCmd := NewAppCmd()
	assert.NotNil(t, appCmd)
	subCommands := appCmd.Commands()
	assert.NotNil(t, subCommands)

	found := false
	for _, subCmd := range subCommands {
		if subCmd.Name() == "list" {
			found = true
		}
	}
	assert.True(t, found, "subcommand list not registered in appCmd")
}
