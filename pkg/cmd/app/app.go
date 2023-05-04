// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/printer"
	appTypes "github.com/tsuru/tsuru/types/app"
	quotaTypes "github.com/tsuru/tsuru/types/quota"
	volumeTypes "github.com/tsuru/tsuru/types/volume"
	"k8s.io/apimachinery/pkg/api/resource"
)

func NewAppCmd() *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "app is a runnable application running on Tsuru",
	}
	appCmd.AddCommand(newAppInfoCmd())
	appCmd.AddCommand(appListCmd)
	return appCmd
}

func completeAppNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	apps, _, err := api.APIClientSingleton().Client.AppApi.AppList(cmd.Context(), &tsuru.AppListOpts{
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

	apps, httpResponse, err := api.APIClientSingleton().Client.AppApi.AppList(cmd.Context(), &tsuru.AppListOpts{
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
}

type app struct {
	IP          string
	CName       []string
	Name        string
	Provisioner string
	Cluster     string
	Platform    string
	Repository  string
	Teams       []string
	Units       []unit
	Owner       string
	TeamOwner   string
	Deploys     uint
	Pool        string
	Description string
	Lock        lock
	Quota       quotaTypes.Quota
	Plan        appTypes.Plan
	Router      string
	RouterOpts  map[string]string
	Tags        []string
	Error       string
	Routers     []appTypes.AppRouter
	AutoScale   []tsuru.AutoScaleSpec

	InternalAddresses    []appInternalAddress
	UnitsMetrics         []unitMetrics
	VolumeBinds          []volumeTypes.VolumeBind
	ServiceInstanceBinds []serviceInstanceBind
}

type serviceInstanceBind struct {
	Service  string
	Instance string
	Plan     string
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

type unit struct {
	ID           string
	IP           string
	InternalIP   string
	Status       string
	StatusReason string
	ProcessName  string
	Address      *url.URL
	Addresses    []url.URL
	Version      int
	Routable     *bool
	Ready        *bool
	Restarts     *int
	CreatedAt    *time.Time
}

func (u *unit) Host() string {
	address := ""
	if len(u.Addresses) > 0 {
		address = u.Addresses[0].Host
	} else if u.Address != nil {
		address = u.Address.Host
	} else if u.IP != "" {
		return u.IP
	}
	if address == "" {
		return address
	}

	host, _, _ := net.SplitHostPort(address)
	return host

}

func (u *unit) ReadyAndStatus() string {
	if u.Ready != nil && *u.Ready {
		return "ready"
	}

	if u.StatusReason != "" {
		return u.Status + " (" + u.StatusReason + ")"
	}

	return u.Status
}

func (u *unit) Port() string {
	if len(u.Addresses) == 0 {
		if u.Address == nil {
			return ""
		}
		_, port, _ := net.SplitHostPort(u.Address.Host)
		return port
	}

	ports := []string{}
	for _, addr := range u.Addresses {
		_, port, _ := net.SplitHostPort(addr.Host)
		ports = append(ports, port)
	}
	return strings.Join(ports, ", ")
}

func (a *app) QuotaString() string {
	var limit strings.Builder
	if a.Quota.IsUnlimited() {
		limit.WriteString("unlimited")
	} else {
		fmt.Fprintf(&limit, "%d units", a.Quota.Limit)
	}
	return fmt.Sprintf("%d/%s", a.Quota.InUse, limit.String())
}

func (a *app) TeamList() string {
	teams := []string{}
	if a.TeamOwner != "" {
		teams = append(teams, a.TeamOwner+" (owner)")
	}

	for _, t := range a.Teams {
		if t != a.TeamOwner {
			teams = append(teams, t)
		}
	}

	return strings.Join(teams, ", ")

}

func (a *app) InternalAddr() string {

	addrs := []string{}
	for _, a := range a.InternalAddresses {
		if a.Protocol == "UDP" {
			addrs = append(addrs, fmt.Sprintf("%s:%d (UDP)", a.Domain, a.Port))
		} else {
			addrs = append(addrs, fmt.Sprintf("%s:%d", a.Domain, a.Port))
		}
	}

	return strings.Join(addrs, ", ")
}

func (a *app) Addr() string {
	var allAddrs []string
	for _, cname := range a.CName {
		if cname != "" {
			allAddrs = append(allAddrs, cname+" (cname)")
		}
	}
	if len(a.Routers) == 0 {
		if a.IP != "" {
			allAddrs = append(allAddrs, a.IP)
		}
	} else {
		for _, r := range a.Routers {
			if len(r.Addresses) > 0 {
				sort.Strings(r.Addresses)
				allAddrs = append(allAddrs, r.Addresses...)
			} else if r.Address != "" {
				allAddrs = append(allAddrs, r.Address)
			}
		}
	}
	return strings.Join(allAddrs, ", ")
}

func (a *app) TagList() string {
	return strings.Join(a.Tags, ", ")
}

func (a *app) GetRouterOpts() string {
	var kv []string
	for k, v := range a.RouterOpts {
		kv = append(kv, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(kv)
	return strings.Join(kv, ", ")
}

func (a *app) SimpleServicesView() string {
	sibs := make([]serviceInstanceBind, len(a.ServiceInstanceBinds))
	copy(sibs, a.ServiceInstanceBinds)

	sort.Slice(sibs, func(i, j int) bool {
		if sibs[i].Service < sibs[j].Service {
			return true
		}
		if sibs[i].Service > sibs[j].Service {
			return false
		}
		return sibs[i].Instance < sibs[j].Instance
	})
	pairs := []string{}
	for _, b := range sibs {
		pairs = append(pairs, b.Service+"/"+b.Instance)
	}

	return strings.Join(pairs, ", ")
}

func memoryValue(q string) string {
	var memory string
	qt, err := resource.ParseQuantity(q)
	if err == nil {
		memory = fmt.Sprintf("%vMi", qt.Value()/(1024*1024))

	}
	return memory
}
