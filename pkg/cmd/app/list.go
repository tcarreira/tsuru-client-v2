package app

import (
	"io"
	"net/http"
	"os"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/printer"
)

func newAppListCmd() *cobra.Command {
	appListCmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "list apps",
		Long: `Lists all apps that you have access to. App access is controlled by teams.
If your team has access to an app, then you have access to it.
Flags can be used to filter the list of applications.`,
		Example: `$ tsuru app list
$ tsuru app list -n my
$ tsuru app list --status error`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appListCmdRun(cmd, args, api.APIClientSingleton(), os.Stdout)
		},
	}

	appListCmd.Flags().StringP("name", "n", "", "filter applications by name")
	appListCmd.Flags().StringP("pool", "o", "", "filter applications by pool")
	appListCmd.Flags().StringP("status", "s", "", "filter applications by unit status. Accepts multiple values separated by commas. Possible values can be: building, created, starting, error, started, stopped, asleep")
	appListCmd.Flags().StringP("platform", "p", "", "filter applications by platform")
	appListCmd.Flags().StringP("team", "t", "", "filter applications by team owner")
	appListCmd.Flags().StringP("user", "u", "", "filter applications by owner")
	appListCmd.Flags().BoolP("locked", "l", false, "filter applications by lock status")
	appListCmd.Flags().BoolP("simplified", "q", false, "display only applications name")
	appListCmd.Flags().Bool("json", false, "display applications in JSON format")
	appListCmd.Flags().StringSliceP("tag", "g", []string{}, "filter applications by tag. Can be used multiple times")

	return appListCmd
}
func appListCmdRun(cmd *cobra.Command, args []string, apiClient *api.APIClient, out io.Writer) error {
	cmd.SilenceUsage = true

	apps, httpResponse, err := apiClient.Client.AppApi.AppList(cmd.Context(), &tsuru.AppListOpts{
		Simplified: optional.NewBool(false),
		Name:       optional.NewString(cmd.Flag("name").Value.String()),
		Platform:   optional.NewString(cmd.Flag("platform").Value.String()),
		TeamOwner:  optional.NewString(cmd.Flag("team").Value.String()),
		Locked:     optional.NewBool(cmd.Flag("locked").Value.String() == "true"),
		Pool:       optional.NewString(cmd.Flag("pool").Value.String()),
		Status:     optional.NewString(cmd.Flag("status").Value.String()),
		Owner: optional.NewString(func() string {
			userFlag := cmd.Flag("user").Value.String()
			if userFlag == "me" {
				user, _, _ := apiClient.Client.UserApi.UserGet(cmd.Context())
				return user.Email
			}
			return userFlag
		}()),
		Tag: optional.NewInterface(cmd.Flag("tag").Value.String()),
	})
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNoContent {
			return nil
		}
		return err
	}

	format := "table"
	if cmd.Flag("json").Value.String() == "true" {
		format = "json"
	}
	return printAppList(out, printer.FormatAs(format), cmd.Flag("simplified").Value.String() == "true", apps)
}

func printAppList(out io.Writer, format printer.OutputType, simplified bool, apps []tsuru.MiniApp) error {
	if format == printer.JSON {
		return printer.PrintPrettyJSON(out, apps)
	}
	return nil
}
