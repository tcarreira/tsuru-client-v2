// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func iterateCmdTreeAndRemoveRun(t *testing.T, cmd *cobra.Command, cmdPath []string, cmdPathChan chan []string) {
	if len(cmd.Commands()) == 0 {
		cmdPathChan <- cmdPath
	}
	for _, c := range cmd.Commands() {
		c.RunE = nil
		c.Run = nil
		newCmdPath := make([]string, len(cmdPath))
		copy(newCmdPath, cmdPath)
		newCmdPath = append(newCmdPath, c.Name())
		iterateCmdTreeAndRemoveRun(t, c, newCmdPath, cmdPathChan)
	}
}

func TestOverridingFlags(t *testing.T) {
	newRootCmd().SetOutput(io.Discard)

	cmdPathChan := make(chan []string)
	go func() {
		iterateCmdTreeAndRemoveRun(t, newRootCmd(), []string{}, cmdPathChan)
		close(cmdPathChan)
	}()

	for cmdPath := range cmdPathChan {
		t.Run(strings.Join(cmdPath, "_"), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic: %v", r)
				}
			}()

			newRootCmd().SetArgs(cmdPath)
			newRootCmd().Execute()
		})
	}
}

func TestParseEnvVariables(t *testing.T) {
	func() { // initConfig() needs some Target
		envName := "TSURU_TARGET"
		if oldEnv, ok := os.LookupEnv(envName); ok {
			defer os.Setenv(envName, oldEnv)
		}
		os.Setenv(envName, "xxx")
		SetupTsuruClientSingleton()
		os.Unsetenv(envName)
	}()

	t.Run("string envs", func(t *testing.T) {
		for _, test := range []struct {
			viperEnvName string
			envName      string
		}{
			{"token", "TSURU_TOKEN"},
			{"target", "TSURU_TARGET"},
			{"auth-schema", "TSURU_AUTH_SCHEMA"},
		} {
			func() {
				if oldEnv, ok := os.LookupEnv(test.envName); ok {
					defer os.Setenv(test.envName, oldEnv)
				}
				os.Setenv(test.envName, "ABCDEFGH")
				assert.Equal(t, "ABCDEFGH", viper.GetString(test.viperEnvName))
				os.Unsetenv(test.envName)
			}()
		}
	})

	t.Run("Int envs", func(t *testing.T) {
		for _, test := range []struct {
			viperEnvName string
			envName      string
		}{
			{"verbosity", "TSURU_VERBOSITY"},
		} {
			func() {
				if oldEnv, ok := os.LookupEnv(test.envName); ok {
					defer os.Setenv(test.envName, oldEnv)
				}
				os.Setenv(test.envName, "123")
				assert.Equal(t, 123, viper.GetInt(test.viperEnvName))
				os.Unsetenv(test.envName)
			}()
		}
	})

	t.Run("Bool envs", func(t *testing.T) {
		for _, test := range []struct {
			viperEnvName string
			envName      string
		}{
			{"insecure-skip-verify", "TSURU_INSECURE_SKIP_VERIFY"},
		} {
			func() {
				if oldEnv, ok := os.LookupEnv(test.envName); ok {
					defer os.Setenv(test.envName, oldEnv)
				}
				os.Setenv(test.envName, "t")
				assert.Equal(t, true, viper.GetBool(test.viperEnvName))
				os.Unsetenv(test.envName)
			}()
		}
	})
}

func TestRunTsuruPluginOrHelp(t *testing.T) {
	t.Run("with no args", func(t *testing.T) {
		cmd := newRootCmd()
		args := []string{}
		tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
		err := runTsuruPluginOrHelp(cmd, args, tsuruCtx)
		assert.NoError(t, err)
		assert.Contains(t, tsuruCtx.Stdout.(*strings.Builder).String(), "A command-line interface for interacting with tsuru")
	})

	t.Run("with args", func(t *testing.T) {
		cmd := &cobra.Command{}
		args := []string{"arg1", "arg2"}
		tsuruCtx := tsuructx.TsuruContextWithConfig(nil)
		err := runTsuruPluginOrHelp(cmd, args, tsuruCtx)
		assert.NoError(t, err)
		assert.Equal(t, "This would the tsuru-plugin: arg1 arg2\nNot implemented yet.\n", tsuruCtx.Stdout.(*strings.Builder).String())
	})
}
