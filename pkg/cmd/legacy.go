package cmd

import (
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
