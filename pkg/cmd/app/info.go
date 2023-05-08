// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/tsuru/tablecli"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/parser"
	"github.com/tsuru/tsuru-client/internal/printer"
	"github.com/tsuru/tsuru-client/pkg/cmd/plan"
	"github.com/tsuru/tsuru-client/pkg/cmd/router"
	"k8s.io/apimachinery/pkg/api/resource"

	appTypes "github.com/tsuru/tsuru/types/app"
	volumeTypes "github.com/tsuru/tsuru/types/volume"
)

const (
	simplifiedFormat = `{{ if .Error -}}
Error: {{ .Error }}
{{ end -}}
Application: {{.Name}}
{{- if .Description }}
Description: {{.Description}}
{{- end }}
{{- if .TagList }}
Tags: {{.TagList}}
{{- end }}
Created by: {{.Owner}}
Platform: {{.Platform}}
Plan: {{ .Plan.Name }}
Pool: {{.Pool}} ({{ .Provisioner }}{{ if .Cluster}} | cluster: {{ .Cluster }}{{end}})
{{if not .Routers -}}
Router:{{if .Router}} {{.Router}}{{if .RouterOpts}} ({{.GetRouterOpts}}){{end}}{{end}}
{{end -}}
Teams: {{.TeamList}}
{{- if .InternalAddr }}
Cluster Internal Addresses: {{.InternalAddr}}
{{- end }}
{{- if .Addr }}
Cluster External Addresses: {{.Addr}}
{{- end }}
{{- if .SimpleServicesView }}
Bound Services: {{ .SimpleServicesView }}
{{- end }}
`
	fullFormat = `{{ if .Error -}}
Error: {{ .Error }}
{{ end -}}
Application: {{.Name}}
{{- if .Description }}
Description: {{.Description}}
{{- end }}
{{- if .TagList }}
Tags: {{.TagList}}
{{- end }}
Platform: {{.Platform}}
{{ if .Provisioner -}}
Provisioner: {{ .Provisioner }}
{{ end -}}
{{if not .Routers -}}
Router:{{if .Router}} {{.Router}}{{if .RouterOpts}} ({{.GetRouterOpts}}){{end}}{{end}}
{{end -}}
Teams: {{.TeamList}}
External Addresses: {{.Addr}}
Created by: {{.Owner}}
Deploys: {{.Deploys}}
{{if .Cluster -}}
Cluster: {{ .Cluster }}
{{ end -}}
Pool:{{if .Pool}} {{.Pool}}{{end}}{{if .Lock.Locked}}
{{.Lock.String}}{{end}}
Quota: {{ .QuotaString }}
`
)

func newAppInfoCmd() *cobra.Command {
	appInfoCmd := &cobra.Command{
		Use:   "info [flags] [app]",
		Short: "shows information about a specific app",
		Long: `shows information about a specific app.
Its name, platform, state (and its units), address, etc.
You need to be a member of a team that has access to the app to be able to see information about it.`,
		Example: `$ tsuru app info myapp
$ tsuru app info -a myapp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printAppInfo(cmd, args, api.APIClientSingleton(), os.Stdout)
		},
		ValidArgsFunction: completeAppNames,
	}

	appInfoCmd.Flags().StringP("app", "a", "", "The name of the app")
	// appInfoCmd.Flags().MarkDeprecated("app", "please use the argument instead")
	appInfoCmd.Flags().MarkHidden("app")
	appInfoCmd.Flags().BoolP("simplified", "s", false, "Show simplified view of app")
	appInfoCmd.Flags().Bool("json", false, "Show JSON view of app")
	return appInfoCmd
}

func printAppInfo(cmd *cobra.Command, args []string, apiClient *api.APIClient, out io.Writer) error {
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

	request, err := apiClient.NewRequest("GET", "/apps/"+appName, nil)
	if err != nil {
		return err
	}
	httpResponse, err := apiClient.RawHTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusNoContent {
		return fmt.Errorf("app %q not found", appName)
	}

	var a app
	err = json.NewDecoder(httpResponse.Body).Decode(&a)
	if err != nil {
		return err
	}

	format := "table"
	if cmd.Flag("json").Value.String() == "true" {
		format = "json"
	}
	return a.PrintInfo(out, printer.FormatAs(format), cmd.Flag("simplified").Value.String() == "true")
}

func (a *app) PrintInfo(out io.Writer, format printer.OutputType, simplified bool) error {
	if format == printer.JSON {
		return printer.PrintPrettyJSON(out, a)
	}

	httpTemplate := fullFormat
	if simplified {
		httpTemplate = simplifiedFormat
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("app").Parse(httpTemplate))

	if simplified {
		renderUnitsSummary(&buf, a.Units, a.UnitsMetrics, a.Provisioner)
	} else {
		renderUnits(&buf, a.Units, a.UnitsMetrics, a.Provisioner)
	}

	internalAddressesTable := tablecli.NewTable()
	internalAddressesTable.Headers = []string{"Domain", "Port", "Process", "Version"}
	for _, internalAddress := range a.InternalAddresses {
		internalAddressesTable.AddRow([]string{
			internalAddress.Domain,
			strconv.Itoa(internalAddress.Port) + "/" + internalAddress.Protocol,
			internalAddress.Process,
			internalAddress.Version,
		})
	}

	if !simplified {
		renderServiceInstanceBinds(&buf, a.ServiceInstanceBinds)
	}

	autoScaleTable := tablecli.NewTable()
	autoScaleTable.Headers = tablecli.Row([]string{"Process", "Min", "Max", "Target CPU"})
	for _, as := range a.AutoScale {
		cpu := parser.CPUValue(as.AverageCPU)
		autoScaleTable.AddRow(tablecli.Row([]string{
			fmt.Sprintf("%s (v%d)", as.Process, as.Version),
			strconv.Itoa(int(as.MinUnits)),
			strconv.Itoa(int(as.MaxUnits)),
			cpu,
		}))
	}

	if autoScaleTable.Rows() > 0 {
		buf.WriteString("\n")
		buf.WriteString("Auto Scale:\n")
		buf.WriteString(autoScaleTable.String())
	}

	if !simplified && (a.Plan.Memory != 0 || a.Plan.Swap != 0 || a.Plan.CpuShare != 0) {
		buf.WriteString("\n")
		buf.WriteString("App Plan:\n")
		buf.WriteString(plan.RenderPlans([]appTypes.Plan{a.Plan}, false, false))
	}
	if !simplified && internalAddressesTable.Rows() > 0 {
		buf.WriteString("\n")
		buf.WriteString("Cluster internal addresses:\n")
		buf.WriteString(internalAddressesTable.String())
	}
	if !simplified && len(a.Routers) > 0 {
		buf.WriteString("\n")
		if a.Provisioner == "kubernetes" {
			buf.WriteString("Cluster external addresses:\n")
			router.RenderRouters(a.Routers, &buf, "Router")
		} else {
			buf.WriteString("Routers:\n")
			router.RenderRouters(a.Routers, &buf, "Name")
		}
	}

	renderVolumeBinds(&buf, a.VolumeBinds)

	var tplBuffer bytes.Buffer
	err := tmpl.Execute(&tplBuffer, a)
	fmt.Fprintln(out, tplBuffer.String()+buf.String())
	return err
}

type lock struct {
	Locked      bool
	Reason      string
	Owner       string
	AcquireDate time.Time
}

func renderUnitsSummary(buf *bytes.Buffer, units []unit, metrics []unitMetrics, provisioner string) {
	type unitsKey struct {
		process  string
		version  int
		routable bool
	}
	groupedUnits := map[unitsKey][]unit{}
	for _, u := range units {
		routable := false
		if u.Routable != nil {
			routable = *u.Routable
		}
		key := unitsKey{process: u.ProcessName, version: u.Version, routable: routable}
		groupedUnits[key] = append(groupedUnits[key], u)
	}
	keys := make([]unitsKey, 0, len(groupedUnits))
	for key := range groupedUnits {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].version == keys[j].version {
			return keys[i].process < keys[j].process
		}
		return keys[i].version < keys[j].version
	})
	var titles []string
	if provisioner == "kubernetes" {
		titles = []string{"Process", "Ready", "Restarts", "Avg CPU (abs)", "Avg Memory"}
	} else {
		titles = []string{"Process", "Units"}
	}
	unitsTable := tablecli.NewTable()
	tablecli.TableConfig.ForceWrap = false
	unitsTable.Headers = tablecli.Row(titles)

	fmt.Fprintf(buf, "Units: %d\n", len(units))

	if len(units) == 0 {
		return
	}
	mapUnitMetrics := map[string]unitMetrics{}
	for _, unitMetric := range metrics {
		mapUnitMetrics[unitMetric.ID] = unitMetric
	}

	for _, key := range keys {
		summaryTitle := key.process
		if key.version > 0 {
			summaryTitle = fmt.Sprintf("%s (v%d)", key.process, key.version)
		}

		summaryUnits := groupedUnits[key]

		if !key.routable && provisioner == "kubernetes" {
			summaryTitle = summaryTitle + " (unroutable)"
		}

		readyUnits := 0
		restarts := 0
		cpuTotal := resource.NewQuantity(0, resource.DecimalSI)
		memoryTotal := resource.NewQuantity(0, resource.BinarySI)

		for _, unit := range summaryUnits {
			if unit.Ready != nil && *unit.Ready {
				readyUnits += 1
			}

			if unit.Restarts != nil {
				restarts += *unit.Restarts
			}

			unitMetric := mapUnitMetrics[unit.ID]
			qt, err := resource.ParseQuantity(unitMetric.CPU)
			if err == nil {
				cpuTotal.Add(qt)
			}
			qt, err = resource.ParseQuantity(unitMetric.Memory)
			if err == nil {
				memoryTotal.Add(qt)
			}
		}

		if provisioner == "kubernetes" {
			unitsTable.AddRow(tablecli.Row{
				summaryTitle,
				fmt.Sprintf("%d/%d", readyUnits, len(summaryUnits)),
				fmt.Sprintf("%d", restarts),
				fmt.Sprintf("%d%%", cpuTotal.MilliValue()/int64(10)/int64(len(summaryUnits))),
				fmt.Sprintf("%vMi", memoryTotal.Value()/int64(1024*1024)/int64(len(summaryUnits))),
			})
		} else {
			unitsTable.AddRow(tablecli.Row{
				summaryTitle,
				fmt.Sprintf("%d", len(summaryUnits)),
			})
		}
	}
	buf.WriteString(unitsTable.String())
}

func renderUnits(buf *bytes.Buffer, units []unit, metrics []unitMetrics, provisioner string) {
	type unitsKey struct {
		process  string
		version  int
		routable bool
	}
	groupedUnits := map[unitsKey][]unit{}
	for _, u := range units {
		routable := false
		if u.Routable != nil {
			routable = *u.Routable
		}
		key := unitsKey{process: u.ProcessName, version: u.Version, routable: routable}
		groupedUnits[key] = append(groupedUnits[key], u)
	}
	keys := make([]unitsKey, 0, len(groupedUnits))
	for key := range groupedUnits {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].version == keys[j].version {
			return keys[i].process < keys[j].process
		}
		return keys[i].version < keys[j].version
	})

	var titles []string
	if provisioner == "kubernetes" {
		titles = []string{"Name", "Host", "Status", "Restarts", "Age", "CPU", "Memory"}
	} else {
		titles = []string{"Name", "Status", "Host", "Port"}
	}
	mapUnitMetrics := map[string]unitMetrics{}
	for _, unitMetric := range metrics {
		mapUnitMetrics[unitMetric.ID] = unitMetric
	}

	for _, key := range keys {
		units := groupedUnits[key]
		unitsTable := tablecli.NewTable()
		tablecli.TableConfig.ForceWrap = false
		unitsTable.Headers = tablecli.Row(titles)
		sort.Slice(units, func(i, j int) bool {
			return units[i].ID < units[j].ID
		})
		for _, unit := range units {
			if unit.ID == "" {
				continue
			}
			var row tablecli.Row
			if provisioner == "kubernetes" {
				row = tablecli.Row{
					unit.ID,
					unit.Host(),
					unit.ReadyAndStatus(),
					parser.IntValue(unit.Restarts),
					parser.TranslateTimestampSince(unit.CreatedAt),
					parser.CPUValue(mapUnitMetrics[unit.ID].CPU),
					memoryValue(mapUnitMetrics[unit.ID].Memory),
				}
			} else {
				row = tablecli.Row{
					parser.ShortID(unit.ID),
					unit.Status,
					unit.Host(),
					unit.Port(),
				}
			}

			unitsTable.AddRow(row)
		}
		if unitsTable.Rows() > 0 {
			unitsTable.SortByColumn(2)
			buf.WriteString("\n")
			groupLabel := ""
			if key.process != "" {
				groupLabel = fmt.Sprintf(" [process %s]", key.process)
			}
			if key.version != 0 {
				groupLabel = fmt.Sprintf("%s [version %d]", groupLabel, key.version)
			}
			if key.routable {
				groupLabel = fmt.Sprintf("%s [routable]", groupLabel)
			}
			buf.WriteString(fmt.Sprintf("Units%s: %d\n", groupLabel, unitsTable.Rows()))
			buf.WriteString(unitsTable.String())
		}
	}
}

func renderServiceInstanceBinds(w io.Writer, binds []serviceInstanceBind) {
	sibs := make([]serviceInstanceBind, len(binds))
	copy(sibs, binds)

	sort.Slice(sibs, func(i, j int) bool {
		if sibs[i].Service < sibs[j].Service {
			return true
		}
		if sibs[i].Service > sibs[j].Service {
			return false
		}
		return sibs[i].Instance < sibs[j].Instance
	})

	type instanceAndPlan struct {
		Instance string
		Plan     string
	}

	instancesByService := map[string][]instanceAndPlan{}
	for _, sib := range sibs {
		instancesByService[sib.Service] = append(instancesByService[sib.Service], instanceAndPlan{
			Instance: sib.Instance,
			Plan:     sib.Plan,
		})
	}

	var services []string
	for _, sib := range sibs {
		if len(services) > 0 && services[len(services)-1] == sib.Service {
			continue
		}
		services = append(services, sib.Service)
	}

	table := tablecli.NewTable()
	table.Headers = []string{"Service", "Instance (Plan)"}

	for _, s := range services {
		var sb strings.Builder
		for i, inst := range instancesByService[s] {
			sb.WriteString(inst.Instance)
			if inst.Plan != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", inst.Plan))
			}

			if i < len(instancesByService[s])-1 {
				sb.WriteString("\n")
			}
		}
		table.AddRow([]string{s, sb.String()})
	}

	if table.Rows() > 0 {
		fmt.Fprintf(w, "\nService instances: %d\n", table.Rows())
		fmt.Fprint(w, table.String())
	}
}

func renderVolumeBinds(w io.Writer, binds []volumeTypes.VolumeBind) {
	table := tablecli.NewTable()
	table.Headers = tablecli.Row([]string{"Name", "MountPoint", "Mode"})
	table.LineSeparator = true

	for _, b := range binds {
		mode := "rw"
		if b.ReadOnly {
			mode = "ro"
		}
		table.AddRow(tablecli.Row([]string{b.ID.Volume, b.ID.MountPoint, mode}))
	}

	if table.Rows() > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Volumes:", table.Rows())
		fmt.Fprint(w, table.String())
	}
}
