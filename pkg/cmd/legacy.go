// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	tsuruV1Config "github.com/tsuru/tsuru-client/tsuru/config"
	tsuruCmd "github.com/tsuru/tsuru/cmd"
)

func newLegacyCommand() *cobra.Command {
	legacyCmd := &cobra.Command{
		Use:   "legacy",
		Short: "legacy is the previous version of tsuru cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLegacyCommand(args)
		},
		Args:               cobra.MinimumNArgs(0),
		DisableFlagParsing: true,
	}
	return legacyCmd
}

func recoverCmdPanicExitError(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(*tsuruCmd.PanicExitError); ok {
			if e.Code > 0 {
				*err = e
			}
			return
		}
		panic(r)
	}
}

func runLegacyCommand(args []string) error {
	var err error
	defer recoverCmdPanicExitError(&err)

	versionForLegacy := strings.TrimLeft(version.Version, "v") + "-legacy-plugin"
	if version.Version == "dev" {
		versionForLegacy = "dev"
	}

	m := tsuruV1Config.BuildManager("tsuru-legacy", versionForLegacy)
	m.Run(args)
	return err
}

type cmdNode struct {
	command  *cobra.Command
	children map[string]*cmdNode
}

func (n *cmdNode) addChild(c *cobra.Command) {
	if n.children == nil {
		n.children = make(map[string]*cmdNode)
	}
	n.children[c.Name()] = &cmdNode{command: c}
	for _, sub := range c.Commands() {
		n.children[c.Name()].addChild(sub)
	}
}

func addMissingLegacyCommands(rootCmd *cobra.Command) []*cobra.Command {
	tree := &cmdNode{command: rootCmd}
	for _, c := range rootCmd.Commands() {
		tree.addChild(c)
	}

	legacyCmd := tsuruV1Config.BuildManager("tsuru-legacy", "")
	for cmdName, v1Cmd := range legacyCmd.Commands {
		curr := tree
		parts := strings.Split(strings.ReplaceAll(cmdName, "-", " "), " ")
		for i, part := range parts {
			if _, ok := curr.children[part]; !ok {
				short := v1Cmd.Info().Usage
				if len(short) > 50 {
					short = short[:50] + "..."
				}
				newCmd := &cobra.Command{
					Use:                part,
					Short:              "[legacy] " + short,
					DisableFlagParsing: true,
				}
				curr.addChild(newCmd)
				curr.command.AddCommand(newCmd)
			}
			curr = curr.children[part]

			if i == len(parts)-1 {
				curr.command.RunE = func(cmd *cobra.Command, args []string) error {
					return fmt.Errorf("aqui %s (%s)", cmdName, v1Cmd.Info().Name)
				}
			}
		}
	}

	return nil
}
