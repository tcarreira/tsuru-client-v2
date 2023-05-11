package app

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/api"
)

// ShellToContainerCmd
func newAppShellCmd() *cobra.Command {
	appShellCmd := &cobra.Command{
		Use:   "shell [FLAGS] APP [UNIT]",
		Short: "run shell inside an app unit",
		Long: `Opens a remote shell inside a unit, using the API server as a proxy. You
can access an app unit just giving app name, or specifying the id of the unit.
You can get the ID of the unit using the "app info" command.
`,
		Example: `$ tsuru app shell myapp
$ tsuru app shell myapp myapp-web-123def-456abc
$ tsuru app shell myapp --isolated`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appShellCmdRun(cmd, args, api.APIClientSingleton(), os.Stdout)
		},
	}

	appShellCmd.Flags().BoolP("isolated", "i", false, "run shell in a new unit")
	return appShellCmd
}

func appShellCmdRun(cmd *cobra.Command, args []string, apiClient *api.APIClient, out io.Writer) error {
	return nil
}
