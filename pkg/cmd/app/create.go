package app

import (
	"github.com/spf13/cobra"
)

func newAppCreateCmd() *cobra.Command {
	appCreateCmd := &cobra.Command{
		Use:   "create [flags] app [platform]",
		Short: "creates a new app",
		Long: `Creates a new app using the given name and platform. For tsuru,
a platform is provisioner dependent. To check the available platforms, use the
command [[tsuru platform list]] and to add a platform use the command [[tsuru platform add]].

In order to create an app, you need to be member of at least one team. All
teams that you are member (see [[tsuru team list]]) will be able to access the
app.

The [[--platform]] parameter is the name of the platform to be used when
creating the app. This will define how tsuru understands and executes your
app. The list of available platforms can be found running [[tsuru platform list]].

The [[--plan]] parameter defines the plan to be used. The plan specifies how
computational resources are allocated to your application. Typically this
means limits for memory and swap usage, and how much cpu share is allocated.
The list of available plans can be found running [[tsuru plan list]].

If this parameter is not informed, tsuru will choose the plan with the
[[default]] flag set to true.

The [[--router]] parameter defines the router to be used. The list of available
routers can be found running [[tsuru router-list]].

If this parameter is not informed, tsuru will choose the router with the
[[default]] flag set to true.

The [[--team]] parameter describes which team is responsible for the created
app, this is only needed if the current user belongs to more than one team, in
which case this parameter will be mandatory.

The [[--pool]] parameter defines which pool your app will be deployed.
This is only needed if you have more than one pool associated with your teams.

The [[--description]] parameter sets a description for your app.
It is an optional parameter, and if its not set the app will only not have a
description associated.

The [[--tag]] parameter sets a tag to your app. You can set multiple [[--tag]] parameters.

The [[--router-opts]] parameter allow passing custom parameters to the router
used by the application's plan. The key and values used depends on the router
implementation.
`,
		Example: `$ tsuru app create myapp
$ tsuru app create myapp python
$ tsuru app create myapp go
$ tsuru app create myapp python --plan small --team myteam
$ tsuru app create myapp python --plan small --team myteam --tag tag1 --tag tag2
`,
		RunE: appCreateCmdRunE,
	}

	appCreateCmd.LocalFlags().StringP("app", "a", "", "The name of the app. Must be unique across tsuru")
	appCreateCmd.LocalFlags().String("platform", "", "The [[--platform]] parameter is the name of the platform to be used when creating the app. This will define how tsuru understands and executes your app. The list of available platforms can be found running [[tsuru platform list]]")
	appCreateCmd.LocalFlags().StringP("description", "d", "", "App description")
	appCreateCmd.LocalFlags().StringP("plan", "p", "", "The plan used to create the app")
	appCreateCmd.LocalFlags().StringP("router", "r", "", "The router used by the app")
	appCreateCmd.LocalFlags().StringP("team", "t", "", "Team owning the app")
	appCreateCmd.LocalFlags().StringP("pool", "o", "", "Pool to deploy your app")
	appCreateCmd.LocalFlags().StringArrayP("tags", "g", []string{}, "App tags")
	appCreateCmd.LocalFlags().StringArray("router-opts", []string{}, "Router options")

	return appCreateCmd
}

func appCreateCmdRunE(cmd *cobra.Command, args []string) error {
	return nil
}
