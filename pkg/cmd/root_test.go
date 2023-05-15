// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
	rootCmd.SetOutput(io.Discard)

	cmdPathChan := make(chan []string)
	go func() {
		iterateCmdTreeAndRemoveRun(t, rootCmd, []string{}, cmdPathChan)
		close(cmdPathChan)
	}()

	for cmdPath := range cmdPathChan {
		t.Run(strings.Join(cmdPath, "_"), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic: %v", r)
				}
			}()

			rootCmd.SetArgs(cmdPath)
			rootCmd.Execute()
		})
	}
}
