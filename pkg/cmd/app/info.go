package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/parser"
	"github.com/tsuru/tsuru-client/internal/printer"
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

	request, err := api.NewRequest("GET", "/apps/"+appName, nil)
	if err != nil {
		return err
	}
	httpResponse, err := api.RawHTTPClient().Do(request)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == 404 {
		return fmt.Errorf("app %q not found", appName)
	}
	var a app
	err = json.NewDecoder(httpResponse.Body).Decode(&a)
	if err != nil {
		return err
	}
	a.prepareAppForPrint()

	printApp := printer.PrintableType{
		SimpleFields: []printer.FieldType{
			{Name: "Name", Value: a.Name},
			{Name: "Description", Value: a.Description},
			{Name: "Deploys", Value: a.Deploys},
			{Name: "Owner", Value: a.Owner},
			{Name: "Platform", Value: a.Platform},
			{Name: "Pool", Value: a.Pool},
			{Name: "Provisioner", Value: a.Provisioner},
			{Name: "Router", Value: a.Router},
			{Name: "Tags", Value: strings.Join(a.Tags, ", ")},
			{Name: "TeamOwner", Value: a.TeamOwner},
			{Name: "Teams", Value: strings.Join(a.Teams, ", ")},
		},
	}
	if len(a.Units) > 0 {
		printApp.DetailedFields = append(printApp.DetailedFields, appUnitsToListType(&a))
	}

	format := "table"
	if cmd.Flag("json").Value.String() == "true" {
		format = "json"
	}
	return printer.PrintInfo(out, printer.FormatAs(format), printApp, nil)
	// return printer.PrintInfo(out, printer.FormatAs(format), a, &printer.TableViewOptions{
	// 	TextTemplate: appInfoTemplate,
	// 	HiddenFields: []string{"CreatedAt", "QuotaJSON", "UnitsMetrics", "Ready"},
	// })
}

func (a *app) prepareAppForPrint() {
	// sort units for printing
	sort.SliceStable(a.Units, func(i, j int) bool {
		if a.Units[i].ProcessName == a.Units[j].ProcessName {
			if a.Units[i].Version == a.Units[j].Version {
				return a.Units[i].CreatedAt.Before(a.Units[j].CreatedAt)
			}
			return a.Units[i].Version < a.Units[j].Version
		}
		return a.Units[i].ProcessName <= a.Units[j].ProcessName
	})

	// Set Calculated fields (for prettier printing)
	mapIDToAppIdx := make(map[string]int, len(a.Units))
	for i, unit := range a.Units {
		mapIDToAppIdx[unit.ID] = i
		a.Units[i].Age = parser.DurationFromTimeWithoutSeconds(unit.CreatedAt, "-")
		if unit.Ready {
			a.Units[i].Status = "ready"
		}
	}

	// Fill cpu and memory on Units
	for _, unitMetrics := range a.UnitsMetrics {
		if idx, ok := mapIDToAppIdx[unitMetrics.ID]; ok {
			a.Units[idx].CPU = parser.CPUValue(unitMetrics.CPU)
			a.Units[idx].Memory = parser.MemoryValue(unitMetrics.Memory)
		}
	}
	a.Quota = fmt.Sprintf("%d/%d", a.QuotaJSON.InUse, a.QuotaJSON.Limit)

}

func appUnitsToListType(a *app) printer.DetailedFieldType {
	return printer.DetailedFieldType{
		Name:   "Units",
		Fields: []string{"Process", "Ver", "Name", "Host", "Status", "Restarts", "Age", "CPU", "Memory"},
		Items: func() []printer.ArrayItemType {
			units := make([]printer.ArrayItemType, len(a.Units))
			for i, u := range a.Units {
				units[i] = printer.ArrayItemType{u.ProcessName, u.Version, u.ID, u.IP, u.Status, u.Restarts, u.Age, u.CPU, u.Memory}
			}
			return units
		}(),
	}
}

func init() {
	appInfoCmd.Flags().StringP("app", "a", "", "The name of the app")
	// appInfoCmd.Flags().MarkDeprecated("app", "please use the argument instead")
	appInfoCmd.Flags().MarkHidden("app")
	appInfoCmd.Flags().BoolP("simplified", "s", false, "Show simplified view of app")
	appInfoCmd.Flags().Bool("json", false, "Show JSON view of app")
}
