// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	tsuruV1Config "github.com/tsuru/tsuru-client/tsuru/config"
	tsuruCmd "github.com/tsuru/tsuru/cmd"
)

func newV1LegacyCmd() *tsuruCmd.Manager {
	versionForLegacy := strings.TrimLeft(version.Version, "v") + "-legacy-plugin"
	if version.Version == "dev" {
		versionForLegacy = "dev"
	}
	return tsuruV1Config.BuildManager("tsuru", versionForLegacy)
}

func newLegacyCommand(v1Cmd *tsuruCmd.Manager) *cobra.Command {
	legacyCmd := &cobra.Command{
		Use:   "legacy",
		Short: "legacy is the previous version of tsuru cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLegacyCommand(v1Cmd, args)
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

func runLegacyCommand(v1Cmd *tsuruCmd.Manager, args []string) error {
	var err error
	defer recoverCmdPanicExitError(&err)

	v1Cmd.Run(args)
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

func addMissingLegacyCommands(rootCmd *cobra.Command, v1RootCmd *tsuruCmd.Manager) {
	// build current commands tree (without legacy commands)
	tree := &cmdNode{command: rootCmd}
	for _, c := range rootCmd.Commands() {
		tree.addChild(c)
	}

	// iterate over legacy commands
	for cmdName, v1Cmd := range v1RootCmd.Commands {
		if cmdName == "help" { // ignore this legacy commands
			continue
		}
		addMissingLegacyCommand(tree, cmdName, v1RootCmd, v1Cmd)
	}
}

func addMissingLegacyCommand(tree *cmdNode, cmdName string, v1RootCmd *tsuruCmd.Manager, v1Cmd tsuruCmd.Command) {
	curr := tree
	parts := strings.Split(strings.ReplaceAll(cmdName, "-", " "), " ")
	for i, part := range parts {
		found := false
		if _, found = curr.children[part]; !found {
			newCmd := &cobra.Command{
				Use:                part,
				Short:              "[v1] " + strings.Join(parts[:i+1], " "),
				DisableFlagParsing: true,
			}
			curr.addChild(newCmd)
			curr.command.AddCommand(newCmd)
		}
		curr = curr.children[part]

		if i == len(parts)-1 && !found {
			curr.command.Short = "[v1] " + strings.Split(v1Cmd.Info().Desc, "\n")[0]
			curr.command.Long = v1Cmd.Info().Usage
			curr.command.SilenceUsage = true
			curr.command.Args = cobra.MinimumNArgs(0)
			curr.command.RunE = func(cmd *cobra.Command, args []string) error {
				return runLegacyCommand(v1RootCmd, append(parts, args...))
			}
		}
	}
}
