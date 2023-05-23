// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
	"golang.org/x/net/websocket"
	"golang.org/x/term"
)

var httpRegexp = regexp.MustCompile(`^http`)

// ShellToContainerCmd
func newAppShellCmd() *cobra.Command {
	appShellCmd := &cobra.Command{
		Use:   "shell APP [UNIT]",
		Short: "run shell inside an app unit",
		Long: `Opens a remote shell inside a unit, using the API server as a proxy. You
can access an app unit just giving app name, or specifying the id of the unit.
You can get the ID of the unit using the "app info" command.
`,
		Example: `$ tsuru app shell myapp
$ tsuru app shell myapp myapp-web-123def-456abc
$ tsuru app shell myapp --isolated`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return appShellCmdRun(cmd, args, tsuructx.GetTsuruContextSingleton(), os.Stdout, os.Stdin)
		},
		Args: cobra.RangeArgs(0, 2),
	}

	appShellCmd.Flags().StringP("app", "a", "", "The name of the app (may be passed as argument)")
	appShellCmd.Flags().StringP("unit", "u", "", "The name of the app's unit (may be passed as argument)")
	appShellCmd.Flags().BoolP("isolated", "i", false, "run shell in a new unit")
	return appShellCmd
}

func appShellCmdRun(cmd *cobra.Command, args []string, tsuruCtx *tsuructx.TsuruContext, out io.Writer, in *os.File) error {
	appName, unitID, err := appNameAndUnitIDFromArgsOrFlags(cmd, args)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true

	qs := make(url.Values)
	qs.Set("isolated", cmd.Flag("isolated").Value.String())
	qs.Set("unit", unitID)
	qs.Set("container_id", unitID)
	width, height := getStdinSize(in)
	qs.Set("width", strconv.Itoa(width))
	qs.Set("height", strconv.Itoa(height))
	if term := os.Getenv("TERM"); term != "" {
		qs.Set("term", term)
	}

	request, err := tsuruCtx.NewRequest("GET", "/apps/"+appName+"/shell", nil)
	if err != nil {
		return err
	}
	request.URL.Scheme = httpRegexp.ReplaceAllString(request.URL.Scheme, "ws")
	reqURLWithoutQuerystring := request.URL.String()
	request.URL.RawQuery = qs.Encode()

	config, err := websocket.NewConfig(request.URL.String(), "ws://localhost")
	if err != nil {
		return err
	}
	config.Header = tsuruCtx.DefaultHeaders()
	config.Dialer = &net.Dialer{}

	/********* wetbsocket does not implement DialWithContext : */
	dialerCancelChan := make(chan struct{})
	config.Dialer.Cancel = dialerCancelChan //lint:ignore SA1019 This is a golang.org/x/net/websocket limitation
	go func() {
		select {
		case <-time.After(5 * time.Second):
			close(dialerCancelChan)
		case <-dialerCancelChan:
		}
	}()
	/********* <- wetbsocket does not implement DialWithContext :( */

	ws, err := websocket.DialConfig(config)
	if err != nil {
		if strings.HasSuffix(err.Error(), "operation was canceled") {
			return fmt.Errorf("timeout connecting to the server: %s", reqURLWithoutQuerystring)
		}
		return err
	}
	close(dialerCancelChan)
	defer ws.Close()
	restoreStdin, err := setupRawStdin(in)
	if err != nil {
		return err
	}
	defer restoreStdin()

	wg := sync.WaitGroup{}
	errChan := make(chan error, 3)
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	wg.Add(1)
	go func() { // handle interrupts
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-interrupt:
				errChan <- fmt.Errorf("interrupted! Closing connection")
				return
			}
		}
	}()

	wg.Add(1)
	go func() { // read from ws and write to stdout
		defer wg.Done()

		_, err1 := copyWithContext(ctx, out, ws)
		if err1 != nil {
			errChan <- err1
			return
		}
	}()

	// wg.Add(1) // leaking this goroutine intentionally. stdin.Read() is blocking
	go func() { // read from stdin and write to ws
		// defer wg.Done()
		defer func() {
			// most important for testing. In real life, this is irrelevant
			time.Sleep(100 * time.Millisecond)
			cancelCtx()
		}()

		_, err1 := copyWithContext(ctx, ws, in)
		if err1 != nil {
			errChan <- err1
			return
		}
	}()

	wg.Wait()

	close(errChan)
	errs := []error{}
	for e := range errChan {
		errs = append(errs, e)
	}

	fmt.Fprintln(out)
	return errors.Join(errs...)
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

func getStdinSize(in *os.File) (width, height int) {
	fd := int(in.Fd())
	if term.IsTerminal(fd) {
		width, height, _ = term.GetSize(fd)
	}
	return
}

func setupRawStdin(in *os.File) (restoreStdin func(), err error) {
	fd := int(in.Fd())
	restoreStdin = func() {}
	if term.IsTerminal(fd) {
		var oldState *term.State
		oldState, err = term.MakeRaw(fd)
		if err != nil {
			return
		}
		restoreStdin = func() {
			term.Restore(fd, oldState)
		}
	}
	return
}

type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	if reflect.ValueOf(src).IsNil() {
		return 0, nil
	}
	return io.Copy(dst, readerFunc(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, io.EOF
		default:
			return src.Read(p)
		}
	}))
}
