package plan

import (
	"fmt"
	"strconv"

	"github.com/tsuru/tablecli"
	appTypes "github.com/tsuru/tsuru/types/app"
	"k8s.io/apimachinery/pkg/api/resource"
)

func RenderPlans(plans []appTypes.Plan, isBytes, showDefaultColumn bool) string {
	table := tablecli.NewTable()
	table.Headers = []string{"Name", "CPU", "Memory"}

	if showDefaultColumn {
		table.Headers = append(table.Headers, "Default")
	}

	for _, p := range plans {
		var cpu, memory string
		if isBytes {
			memory = fmt.Sprintf("%d", p.Memory)
		} else {
			memory = resource.NewQuantity(p.Memory, resource.BinarySI).String()
		}

		if p.Override.CPUMilli != nil {
			cpu = fmt.Sprintf("%g", float64(*p.Override.CPUMilli)/10) + "% (override)"
		} else if p.CPUMilli > 0 {
			cpu = fmt.Sprintf("%g", float64(p.CPUMilli)/10) + "%"
		}

		if p.Override.Memory != nil {
			memory = resource.NewQuantity(*p.Override.Memory, resource.BinarySI).String() + " (override)"
		}

		row := []string{
			p.Name,
			cpu,
			memory,
		}

		if showDefaultColumn {
			row = append(row, strconv.FormatBool(p.Default))
		}
		table.AddRow(row)
	}
	return table.String()
}
