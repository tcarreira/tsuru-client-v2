// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/tsuru/tablecli"
	appTypes "github.com/tsuru/tsuru/types/app"
)

func RenderRouters(routers []appTypes.AppRouter, out io.Writer, idColumn string) {
	table := tablecli.NewTable()
	table.Headers = tablecli.Row([]string{idColumn, "Opts", "Addresses", "Status"})
	table.LineSeparator = true
	for _, r := range routers {
		var optsStr []string
		for k, v := range r.Opts {
			optsStr = append(optsStr, fmt.Sprintf("%s: %s", k, v))
		}
		sort.Strings(optsStr)
		statusStr := r.Status
		if r.StatusDetail != "" {
			statusStr = fmt.Sprintf("%s: %s", statusStr, r.StatusDetail)
		}
		addresses := r.Address
		if len(r.Addresses) > 0 {
			sort.Strings(r.Addresses)
			addresses = strings.Join(r.Addresses, "\n")
		}
		row := tablecli.Row([]string{
			r.Name,
			strings.Join(optsStr, "\n"),
			addresses,
			statusStr,
		})
		table.AddRow(row)
	}
	out.Write(table.Bytes())
}
