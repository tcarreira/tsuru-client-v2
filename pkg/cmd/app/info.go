/*
Copyright © 2023 tsuru-client authors
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
package app

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/printer"
)

var appInfoCmd = &cobra.Command{
	Use:   "info [flags] [app]",
	Short: "shows information about a specific app",
	Long: `shows information about a specific app.
Its name, platform, state (and its units), address, etc.
You need to be a member of a team that has access to the app to be able to see information about it.`,
	Example: `$ tsuru app info myapp
$ tsuru app info -a myapp
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printAppInfo(cmd, args, os.Stdout)
	},
}

func printAppInfo(cmd *cobra.Command, args []string, out io.Writer) error {
	if len(args) == 0 && cmd.Flag("app").Value.String() == "" {
		return fmt.Errorf("no app was provided. Please provide an app name or use the --app flag")
	}
	if len(args) > 0 && cmd.Flag("app").Value.String() != "" {
		return fmt.Errorf("either pass an app name as an argument or use the --app flag, not both")
	}
	cmd.SilenceUsage = true

	appName := cmd.Flag("app").Value.String()
	if appName == "" {
		appName = args[0]
	}

	app, httpResponse, err := api.Client().AppApi.AppGet(cmd.Context(), appName)
	if err != nil {
		return err
	}
	if httpResponse.StatusCode == 404 {
		return fmt.Errorf("app %q not found", appName)
	}

	format := "table"
	if cmd.Flag("json").Value.String() == "true" {
		format = "json"
	}
	w := tabwriter.NewWriter(out, 2, 2, 2, ' ', 0)
	return printer.PrintInfo(w, printer.FormatAs(format), app)
}

func init() {
	appInfoCmd.Flags().StringP("app", "a", "", "The name of the app")

	appCmd.AddCommand(appInfoCmd)
}
