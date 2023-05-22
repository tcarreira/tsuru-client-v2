// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/api"
)

func newAppLogCmd() *cobra.Command {
	appLogCmd := &cobra.Command{
		Use:   "log [FLAGS] APP [UNIT]",
		Short: "shows log entries for an application",
		Long: `Shows log entries for an application. These logs include everything the
application send to stdout and stderr, alongside with logs from tsuru server
(deployments, restarts, etc.)

The [[--lines]] flag is optional and by default its value is 10.

The [[--source]] flag is optional and allows filtering logs by log source
(e.g. application, tsuru api).

The [[--unit]] flag is optional and allows filtering by unit. It's useful if
your application has multiple units and you want logs from a single one.

The [[--follow]] flag is optional and makes the command wait for additional
log output

The [[--no-date]] flag is optional and makes the log output without date.

The [[--no-source]] flag is optional and makes the log output without source
information, useful to very dense logs.
`,
		Example: `$ tsuru app log myapp
$ tsuru app log -l 50 -f myapp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appLogCmdRun(cmd, args, api.APIClientSingleton(), os.Stdout)
		},
	}

	appLogCmd.Flags().StringP("app", "a", "", "The name of the app (may be passed as argument)")
	appLogCmd.Flags().StringP("unit", "u", "", "The log from the given unit (may be passed as argument)")
	appLogCmd.Flags().IntP("lines", "l", 10, "The number of log lines to display")
	appLogCmd.Flags().StringP("source", "s", "", "The log from the given source")
	appLogCmd.Flags().BoolP("follow", "f", false, "Follow logs")
	appLogCmd.Flags().Bool("no-date", false, "No date information")
	appLogCmd.Flags().Bool("no-source", false, "No source information")

	return appLogCmd
}

func appLogCmdRun(cmd *cobra.Command, args []string, apiClient *api.APIClient, out io.Writer) error {
	return nil
}
