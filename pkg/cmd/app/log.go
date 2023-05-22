// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tsuru/tsuru-client/internal/api"
	"github.com/tsuru/tsuru-client/internal/printer"
)

const tLogFmt = "2006-01-02 15:04:05 -0700"

func newAppLogCmd() *cobra.Command {
	appLogCmd := &cobra.Command{
		Use:   "log [FLAGS] APP [UNIT]",
		Short: "shows log entries for an application",
		Long: `Shows log entries for an application. These logs include everything the
application send to stdout and stderr, alongside with logs from tsuru server
(deployments, restarts, etc.)

The [[--lines]] flag is optional and by default its value is 10.

The [[--source]] flag is optional and allows filtering logs by log source
(e.g. application, tsuru api).

The [[--unit]] flag is optional and allows filtering by unit. It's useful if
your application has multiple units and you want logs from a single one.

The [[--follow]] flag is optional and makes the command wait for additional
log output

The [[--no-date]] flag is optional and makes the log output without date.

The [[--no-source]] flag is optional and makes the log output without source
information, useful to very dense logs.
`,
		Example: `$ tsuru app log myapp
$ tsuru app log -l 50 -f myapp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appLogCmdRun(cmd, args, api.APIClientSingleton(), os.Stdout)
		},
	}

	appLogCmd.Flags().StringP("app", "a", "", "The name of the app (may be passed as argument)")
	appLogCmd.Flags().StringP("unit", "u", "", "The log from the given unit (may be passed as argument)")
	appLogCmd.Flags().IntP("lines", "l", 10, "The number of log lines to display")
	appLogCmd.Flags().StringP("source", "s", "", "The log from the given source")
	appLogCmd.Flags().BoolP("follow", "f", false, "Follow logs")
	appLogCmd.Flags().Bool("no-date", false, "No date information")
	appLogCmd.Flags().Bool("no-source", false, "No source information")

	return appLogCmd
}

func appLogCmdRun(cmd *cobra.Command, args []string, apiClient *api.APIClient, out io.Writer) error {
	appName, unitID, err := appNameAndUnitIDFromArgsOrFlags(cmd, args)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true

	request, err := apiClient.NewRequest("GET", "/apps/"+appName+"/log", nil)
	if err != nil {
		return err
	}
	qs := make(url.Values)
	qs.Set("lines", cmd.Flag("lines").Value.String())
	if unitID != "" {
		qs.Set("unit", unitID)
	}
	if cmd.Flag("source").Value.String() != "" {
		qs.Set("source", cmd.Flag("source").Value.String())
	}
	if isFollow, _ := cmd.Flags().GetBool("follow"); isFollow {
		qs.Set("follow", "1")
	}
	request.URL.RawQuery = qs.Encode()
	httpResponse, err := apiClient.RawHTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == http.StatusNoContent {
		return nil
	}

	formatter := logFormatter{
		noDate:   func() bool { v, _ := cmd.Flags().GetBool("no-date"); return v }(),
		noSource: func() bool { v, _ := cmd.Flags().GetBool("no-source"); return v }(),
		localTZ:  apiClient.Opts.LocalTZ,
	}
	dec := json.NewDecoder(httpResponse.Body)
	for {
		err = formatter.Format(out, dec)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(out, "Error: %v", err)
			}
			break
		}
	}
	return nil
}

type logFormatter struct {
	noDate   bool
	noSource bool
	localTZ  *time.Location
}
type log struct {
	Date    time.Time
	Message string
	Source  string
	Unit    string
}

func (f logFormatter) Format(out io.Writer, dec *json.Decoder) error {
	var logs []log
	err := dec.Decode(&logs)
	if err != nil {
		if err == io.EOF {
			return err
		}
		buffered := dec.Buffered()
		bufferedData, _ := io.ReadAll(buffered)
		return fmt.Errorf("unable to parse json: %v: %q", err, string(bufferedData))
	}
	colorify := printer.Colorify{
		DisableColors: viper.IsSet("disable-colors"),
	}
	for _, l := range logs {
		prefix := f.prefix(l)

		if prefix == "" {
			fmt.Fprintf(out, "%s\n", l.Message)
		} else {
			fmt.Fprintf(out, "%s %s\n", colorify.Colorfy(prefix, "blue", "", ""), l.Message)
		}
	}
	return nil
}

func (f logFormatter) prefix(l log) string {
	parts := make([]string, 0, 2)
	if !f.noDate {
		parts = append(parts, l.Date.In(f.localTZ).Format(tLogFmt))
	}
	if !f.noSource {
		if l.Unit != "" {
			parts = append(parts, fmt.Sprintf("[%s][%s]", l.Source, l.Unit))
		} else {
			parts = append(parts, fmt.Sprintf("[%s]", l.Source))
		}
	}
	prefix := strings.Join(parts, " ")
	if prefix != "" {
		prefix = prefix + ":"
	}
	return prefix
}
