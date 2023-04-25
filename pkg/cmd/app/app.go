// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"io"
	"os"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/printer"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "app is a runnable application running on Tsuru",
}

func AppCmd() *cobra.Command {
	return appCmd
}

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
	ValidArgsFunction: completeAppNames,
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
	return printer.PrintInfo(out, printer.FormatAs(format), app)
}

func completeAppNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apps, _, err := api.Client().AppApi.AppList(cmd.Context(), &tsuru.AppListOpts{
		Simplified: optional.NewBool(true),
		Name:       optional.NewString(toComplete),
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []string
	for _, app := range apps {
		names = append(names, app.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	appCmd.PersistentFlags().StringP("app", "a", "", "The name of the app")
	appCmd.PersistentFlags().MarkDeprecated("app", "please use the argument instead")
	appCmd.PersistentFlags().MarkHidden("app")

	appCmd.AddCommand(appInfoCmd)
}
