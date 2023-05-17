// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/api"
	"golang.org/x/net/websocket"
	"golang.org/x/term"
)

var httpRegexp = regexp.MustCompile(`^http`)

// ShellToContainerCmd
func newAppShellCmd() *cobra.Command {
	appShellCmd := &cobra.Command{
		Use:   "shell [FLAGS] APP [UNIT]",
		Short: "run shell inside an app unit",
		Long: `Opens a remote shell inside a unit, using the API server as a proxy. You
can access an app unit just giving app name, or specifying the id of the unit.
You can get the ID of the unit using the "app info" command.
`,
		Example: `$ tsuru app shell myapp
$ tsuru app shell myapp myapp-web-123def-456abc
$ tsuru app shell myapp --isolated`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appShellCmdRun(cmd, args, api.APIClientSingleton(), os.Stdout, os.Stdin)
		},
	}

	appShellCmd.Flags().StringP("app", "a", "", "The name of the app (may be passed as argument)")
	appShellCmd.Flags().StringP("unit", "u", "", "The name of the app's unit (may be passed as argument)")
	appShellCmd.Flags().BoolP("isolated", "i", false, "run shell in a new unit")
	return appShellCmd
}

func appShellCmdRun(cmd *cobra.Command, args []string, apiClient *api.APIClient, out io.Writer, in *os.File) error {
	cmd.SilenceUsage = true

	appName, unitID, err := appNameAndUnitIDFromArgsOrFlags(cmd, args)
	if err != nil {
		return err
	}

	width, height, restoreStdin, err := setupStdin(in)
	if err != nil {
		return err
	}
	defer restoreStdin()

	qs := make(url.Values)
	qs.Set("isolated", cmd.Flag("isolated").Value.String())
	qs.Set("width", strconv.Itoa(width))
	qs.Set("height", strconv.Itoa(height))
	qs.Set("unit", unitID)
	qs.Set("container_id", unitID)
	if term := os.Getenv("TERM"); term != "" {
		qs.Set("term", term)
	}

	request, err := apiClient.NewRequest("GET", "/apps/"+appName+"/shell", nil)
	if err != nil {
		return err
	}
	request.URL.RawQuery = qs.Encode()
	request.URL.Scheme = httpRegexp.ReplaceAllString(request.URL.Scheme, "ws")

	config, err := websocket.NewConfig(request.URL.String(), "ws://localhost")
	if err != nil {
		return err
	}

	config.Header = apiClient.DefaultHeaders()

	ws, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}
	defer ws.Close()

	wg := sync.WaitGroup{}
	errs := make(chan error, 2)
	ctx, cancelCtx := context.WithCancel(context.Background())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	wg.Add(1)
	go func() { // handle interrupts
		defer wg.Done()
		defer cancelCtx()
		for {
			select {
			case <-ctx.Done():
				return
			case <-interrupt:
				errs <- fmt.Errorf("interrupted! Closing connection")
				return
			}
		}
	}()

	wg.Add(1)
	go func() { // read from ws and write to stdout
		defer wg.Done()
		defer cancelCtx()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var buf = make([]byte, 1024)
				n, err := ws.Read(buf)
				if err != nil {
					if err != io.EOF {
						errs <- err
					}
					return
				}
				out.Write(buf[:n])
			}
		}
	}()
	wg.Add(1)
	go func() { // read from stdin and write to ws
		defer wg.Done()
		defer cancelCtx()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				buf := make([]byte, 1024)
				n, err := in.Read(buf)
				if err != nil {
					if err != io.EOF {
						errs <- err
					}
					return
				}
				ws.Write(buf[:n])
			}
		}
	}()

	wg.Wait()
	close(errs)
	return <-errs
}

func appNameAndUnitIDFromArgsOrFlags(cmd *cobra.Command, args []string) (appName, unitID string, err error) {
	appName = cmd.Flag("app").Value.String()
	unitID = cmd.Flag("unit").Value.String()
	switch len(args) {
	case 0:
		if appName == "" {
			return "", "", fmt.Errorf("app name is required")
		}
	case 1:
		if appName == "" {
			appName = args[0]
		} else {
			if unitID != "" {
				return "", "", fmt.Errorf("specify app and unit either by flags or by arguments, not both")
			}
			unitID = args[0]
		}
	case 2:
		if appName != "" || unitID != "" {
			return "", "", fmt.Errorf("specify app and unit either by flags or by arguments, not both")
		}
		appName = args[0]
		unitID = args[1]
	default:
		return "", "", fmt.Errorf("too many arguments")
	}
	return
}

func setupStdin(in *os.File) (width, height int, restoreStdin func(), err error) {
	fd := int(in.Fd())
	restoreStdin = func() {}
	if term.IsTerminal(fd) {
		width, height, _ = term.GetSize(fd)
		var oldState *term.State
		oldState, err = term.MakeRaw(fd)
		if err != nil {
			return
		}
		restoreStdin = func() { term.Restore(fd, oldState) }
		sigChan := make(chan os.Signal, 2)
		go func(c <-chan os.Signal) {
			if _, ok := <-c; ok {
				term.Restore(fd, oldState)
				os.Exit(1)
			}
		}(sigChan)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	}
	return
}
