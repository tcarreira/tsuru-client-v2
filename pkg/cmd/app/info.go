/*
Copyright © 2023 tsuru-client authors
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/parser"
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

	if cmd.Flag("json").Value.String() == "true" {
		appByte, err := json.MarshalIndent(app, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(out, string(appByte))
	} else {
		w := tabwriter.NewWriter(out, 3, 3, 2, ' ', 0)
		defer w.Flush()

		printAppMetadata(w, app)
		fmt.Fprintf(w, "\n\n")
		printAppUnits(w, app)
		printAppServiceInstances(w, app)

	}

	return nil
}

func printAppMetadata(w io.Writer, app tsuru.App) {
	fmt.Fprintf(w, "Application:\t%s\n", app.Name)
	fmt.Fprintf(w, "Description:\t%s\n", app.Description)
	fmt.Fprintf(w, "Tags:\t%s\n", strings.Join(app.Tags, ", "))
	fmt.Fprintf(w, "Platform:\t%s\n", app.Platform)
	fmt.Fprintf(w, "Plan:\t%s (CPU: %s, Memory: %s)\n", app.Plan.Name, parser.CPUMilliToPercent(app.Plan.Cpumilli), parser.MemoryToHuman(app.Plan.Memory))
	fmt.Fprintf(w, "Provisioner:\t%s\n", app.Provisioner)
	fmt.Fprintf(w, "Teams:\t%s (owner)%s\n", app.TeamOwner, strings.Join(append([]string{""}, app.Teams...), ", "))
	fmt.Fprintf(w, "External Addresses:\t%s\n", "TODO (not-implemented!)")
	fmt.Fprintf(w, "Created by:\t%s\n", app.Owner)
	fmt.Fprintf(w, "Deploys:\t%d\n", app.Deploys)
	fmt.Fprintf(w, "Cluster:\t%s\n", app.Cluster)
	fmt.Fprintf(w, "Pool:\t%s\n", app.Pool)
	fmt.Fprintf(w, "Quota:\t%s\n", "TODO (not-implemented!)")
}

func printAppUnits(w io.Writer, app tsuru.App) {
	unitsMapping := map[string]map[int32][]tsuru.Unit{}
	for _, unit := range app.Units {
		if unitsMapping[unit.Processname] == nil {
			unitsMapping[unit.Processname] = map[int32][]tsuru.Unit{}
		}
		unitsMapping[unit.Processname][unit.Version] = append(unitsMapping[unit.Processname][unit.Version], unit)
	}

	first := true
	for process, versions := range unitsMapping {
		for version, units := range versions {
			if !first {
				fmt.Fprintf(w, "\t\t\t\t\t\t\n")
			}
			first = false

			fmt.Fprintf(w, "\t\t\t\t\t\t\rUnits [Process: %s] [version %d]: %d\n", process, version, len(units)) // \t\t\r to ignore column-format on this line
			fmt.Fprintf(w, "NAME\tHOST\tSTATUS\tRESTARTS\tAGE\tCPU\tMEMORY\t\n")
			for _, unit := range units {
				fmt.Fprintf(w, "%s\t%s\t%v\t%d\t%s\t%s\t%s\t\n", unit.Name, unit.Ip, *unit.Routable, *unit.Restarts, parser.DurationFromTimeWithoutSeconds(unit.CreatedAt, "-"), "-", "-")
			}
		}
	}
}

func printAppServiceInstances(w io.Writer, app tsuru.App) {}

func init() {
	appInfoCmd.Flags().StringP("app", "a", "", "The name of the app")

	appCmd.AddCommand(appInfoCmd)
}
