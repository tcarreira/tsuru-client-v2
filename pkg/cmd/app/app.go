// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

var appInfoTemplate = `Name:	{{ .Name }}
Cluster:	{{ .Cluster }}
Cname:	{{ .CName | Join }}
Deploys:	{{ .Deploys }}
Description:	{{ .Description }}
Owner:	{{ .Owner }}
Platform:	{{ .Platform }}
Pool:	{{ .Pool }}
Provisioner:	{{ .Provisioner }}
Router:	{{ .Router }}
Tags:	{{ .Tags | Join }}
TeamOwner:	{{ .TeamOwner }}
Teams:	{{ .Teams | Join }}

{{- if .Units }}

Units:
	PROCESS	VER	NAME	HOST	STATUS	RESTARTS	AGE	CPU	MEMORY
{{- range .Units }}
	{{ .ProcessName }}	{{ .Version }}	{{ .ID }}	{{ .IP }}	{{ .Status }}	{{ .Restarts }}	{{ .Age }}	{{ .CPU }}	{{ .Memory }}{{ end }}
{{- end }}
`

type app struct {
	Cluster     string
	CName       []string
	Deploys     uint
	Description string
	Error       string
	Lock        lock
	Name        string
	Owner       string
	Plan        plan
	Platform    string
	Pool        string
	Provisioner string
	Quota       string `json:"-"`
	QuotaJSON   quota  `json:"quota"`
	Repository  string
	Router      string
	RouterOpts  map[string]string
	Tags        []string
	TeamOwner   string
	Teams       []string

	AutoScale            []tsuru.AutoScaleSpec
	InternalAddresses    []appInternalAddress
	Routers              []appRouter
	ServiceInstanceBinds []serviceInstanceBind
	Units                []unit
	UnitsMetrics         []unitMetrics
	VolumeBinds          []volumeBind
}

type volumeBindID struct {
	App        string
	MountPoint string
	Volume     string
}

type serviceInstanceBind struct {
	Service  string
	Instance string
	Plan     string
}

type volumeBind struct {
	ID       volumeBindID
	ReadOnly bool
}

type appInternalAddress struct {
	Domain   string
	Protocol string
	Port     int
	Version  string
	Process  string
}

type unitMetrics struct {
	ID     string
	CPU    string
	Memory string
}

type appRouter struct {
	Name         string            `json:"name"`
	Opts         map[string]string `json:"opts"`
	Address      string            `json:"address"`
	Addresses    []string          `json:"addresses"`
	Type         string            `json:"type"`
	Status       string            `json:"status,omitempty"`
	StatusDetail string            `json:"status-detail,omitempty"`
}

type unit struct {
	ID string
	IP string
	// InternalIP   string
	Status       string
	StatusReason string
	ProcessName  string
	// Address      *url.URL
	// Addresses    []url.URL
	Version   int
	Routable  bool
	Ready     bool
	Restarts  int
	CreatedAt time.Time
	Age       string
	CPU       string
	Memory    string
}

type lock struct {
	Locked      bool
	Reason      string
	Owner       string
	AcquireDate time.Time
}

type quota struct {
	Limit int `json:"limit"`
	InUse int `json:"inuse"`
}

type plan struct {
	Name   string `json:"name"`
	Memory int64  `json:"memory"`
	Swap   int64  `json:"swap"`
	// CpuShare is DEPRECATED, use CPUMilli instead
	CpuShare int          `json:"cpushare"`
	CPUMilli int          `json:"cpumilli"`
	Default  bool         `json:"default,omitempty"`
	Override planOverride `json:"override,omitempty"`
}

type planOverride struct {
	Memory   int64 `json:"memory"`
	CPUMilli int   `json:"cpumilli"`
}

func AppCmd() *cobra.Command {
	return appCmd
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

var appListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "list apps",
	Long: `Lists all apps that you have access to. App access is controlled by teams.
If your team has access to an app, then you have access to it.
Flags can be used to filter the list of applications.`,
	Example: `$ tsuru app list
$ tsuru app list -n my
$ tsuru app list --status error
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printAppList(cmd, args, os.Stdout)
	},
}

func printAppList(cmd *cobra.Command, args []string, out io.Writer) error {
	cmd.SilenceUsage = true

	apps, httpResponse, err := api.Client().AppApi.AppList(cmd.Context(), &tsuru.AppListOpts{
		Simplified: optional.NewBool(false),
		Name:       optional.NewString(cmd.Flag("name").Value.String()),
		TeamOwner:  optional.NewString(cmd.LocalFlags().Lookup("team").Value.String()),
		Status:     optional.NewString(cmd.Flag("status").Value.String()),
		Locked:     optional.NewBool(cmd.Flag("locked").Value.String() == "true"),
		Owner:      optional.NewString(cmd.Flag("user").Value.String()),
		Platform:   optional.NewString(cmd.Flag("platform").Value.String()),
		Pool:       optional.NewString(cmd.Flag("pool").Value.String()),
		// Tag:        optional.NewString(cmd.Flag("tag").Value.String()), //XXX: fix this
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
	return printer.PrintList(out, printer.FormatAs(format), apps, &printer.TableViewOptions{
		ShowFields: []string{"Name", "Pool", "TeamOwner", "Units"},
		CustomFieldFunc: map[string]printer.CustomFieldFunc{
			"Units": func(data any) string {
				item := data.(tsuru.MiniApp)
				return fmt.Sprintf("%d units", len(item.Units))
			},
		},
	})
}

func init() {
	appListCmd.Flags().StringP("name", "n", "", "Filter applications by name")
	appListCmd.Flags().StringP("pool", "o", "", "Filter applications by pool")
	appListCmd.Flags().StringP("platform", "p", "", "Filter applications by platform")
	appListCmd.LocalFlags().StringP("team", "t", "", "Filter applications by team owner")
	appListCmd.Flags().StringP("user", "u", "", "Filter applications by owner")
	appListCmd.Flags().StringP("status", "s", "", "Filter applications by unit status. Accepts multiple values separated by commas. Possible values can be: building, created, starting, error, started, stopped, asleep")
	appListCmd.Flags().StringSliceP("tag", "g", []string{}, "Filter applications by tag. Can be used multiple times")
	appListCmd.Flags().BoolP("locked", "l", false, "Filter applications by lock status")
	appListCmd.Flags().BoolP("names-only", "q", false, "Display only applications name")

	appCmd.AddCommand(appInfoCmd)
	appCmd.AddCommand(appListCmd)
}
