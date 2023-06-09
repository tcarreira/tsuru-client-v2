// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/v2/internal/parser"
	"github.com/tsuru/tsuru-client/v2/internal/tsuructx"
)

func newAppCreateCmd(tsuruCtx *tsuructx.TsuruContext) *cobra.Command {
	appCreateCmd := &cobra.Command{
		Use:   "create APP [PLATFORM]",
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
$ tsuru app create myapp python --plan small --team myteam --tag tag1 --tag tag2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appCreateRun(tsuruCtx, cmd, args)
		},
		Args: cobra.RangeArgs(0, 2),
	}

	appCreateCmd.Flags().StringP("app", "a", "", "the name of the app. Must be unique across tsuru (may be passed as argument)")
	appCreateCmd.Flags().String("platform", "", "the platform for the app (can be changed later) (may be passed as argument)")
	appCreateCmd.Flags().StringP("description", "d", "", "app description")
	appCreateCmd.Flags().StringP("plan", "p", "", "the plan used to create the app")
	appCreateCmd.Flags().StringP("router", "r", "", "the router used by the app")
	appCreateCmd.Flags().StringP("team", "t", "", "team owning the app")
	appCreateCmd.Flags().StringP("pool", "o", "", "pool to deploy your app")
	appCreateCmd.Flags().StringArrayP("tag", "g", nil, "app tags")
	appCreateCmd.Flags().StringArray("router-opts", nil, "router options")

	return appCreateCmd
}

func appCreateRun(tsuruCtx *tsuructx.TsuruContext, cmd *cobra.Command, args []string) error {
	var appName, platform string
	if len(args) == 0 && cmd.Flag("app").Value.String() == "" {
		return fmt.Errorf("no app was provided. Please provide an app name")
	}
	if len(args) > 0 && cmd.Flag("app").Value.String() != "" {
		return fmt.Errorf("flag --app and argument app cannot be used at the same time")
	}
	cmd.SilenceUsage = true

	if len(args) > 0 {
		appName = args[0]
	}
	if len(args) > 1 {
		platform = args[1]
	}

	if appName != "" && cmd.Flag("app").Value.String() != "" {
		return fmt.Errorf("flag --app and argument app cannot be used at the same time")
	}
	if platform != "" && cmd.Flag("platform").Value.String() != "" {
		return fmt.Errorf("flag --platform and argument platform cannot be used at the same time")
	}

	v := url.Values{}
	v.Set("name", appName)
	v.Set("platform", platform)
	v.Set("description", cmd.Flag("description").Value.String())
	v.Set("plan", cmd.Flag("plan").Value.String())
	v.Set("router", cmd.Flag("router").Value.String())
	v.Set("teamOwner", cmd.Flag("team").Value.String())
	v.Set("pool", cmd.Flag("pool").Value.String())
	if tags, err := cmd.Flags().GetStringArray("tag"); err == nil {
		for _, tag := range tags {
			v.Add("tag", tag)
		}
	}
	if routerOpts, err := cmd.Flags().GetStringArray("router-opts"); err == nil {
		routerOptsMap, err := parser.SliceToMapFlags(routerOpts)
		if err != nil {
			return err
		}
		for key, val := range routerOptsMap {
			v.Add("routeropts."+key, val)
		}
	}

	b := strings.NewReader(v.Encode())
	request, err := tsuruCtx.NewRequest("POST", "/apps", b)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := tsuruCtx.RawHTTPClient().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	data := make(map[string]string)
	err = json.Unmarshal(result, &data)
	if err != nil {
		return err
	}
	fmt.Fprintf(tsuruCtx.Stdout, "App %q has been created!\n", appName)
	fmt.Fprintln(tsuruCtx.Stdout, "Use app info to check the status of the app and its units.")
	return nil
}
