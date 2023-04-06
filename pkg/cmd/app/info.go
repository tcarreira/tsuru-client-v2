/*
Copyright © 2023 tsuru-client authors
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
package app

import (
	"fmt"

	"github.com/spf13/cobra"
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
		if len(args) == 0 && cmd.Flag("app").Value.String() == "" {
			return fmt.Errorf("no app was provided. Please provide an app name or use the --app flag")
		}
		if len(args) > 0 && cmd.Flag("app").Value.String() != "" {
			return fmt.Errorf("either pass an app name as an argument or use the --app flag, not both")
		}

		appName := cmd.Flag("app").Value.String()
		if appName == "" {
			appName = args[0]
		}

		fmt.Println("printing app info for: " + appName)
		return nil
	},
}

func init() {
	appInfoCmd.Flags().StringP("app", "a", "", "The name of the app")

	appCmd.AddCommand(appInfoCmd)
}
