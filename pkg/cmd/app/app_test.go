// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/tsuru/tsuru-client/v2/internal/tsuructx"
)

func TestAppNameAndUnitIDFromArgsOrFlags(t *testing.T) {
	t.Parallel()
	newCmd := func(tsuruCtx *tsuructx.TsuruContext) *cobra.Command {
		newCmd := &cobra.Command{
			Args: cobra.RangeArgs(0, 2),
		}
		newCmd.Flags().StringP("app", "a", "", "app name")
		newCmd.Flags().StringP("unit", "u", "", "unit name")
		return newCmd
	}
	for i, test := range []struct {
		flags      []string
		args       []string
		expectApp  string
		expectUnit string
		err        error
	}{
		{[]string{}, []string{}, "", "", nil},
		{[]string{"-a", "myapp"}, []string{}, "myapp", "", nil},
		{[]string{"-a", "myapp"}, []string{"myunit"}, "myapp", "myunit", nil},
		{[]string{}, []string{"myapp", "myunit"}, "myapp", "myunit", nil},
		{[]string{"-u", "myunit"}, []string{}, "", "myunit", nil},
		{[]string{"-u", "myunit"}, []string{"myapp"}, "myapp", "myunit", nil},
		{[]string{"-a", "myapp", "-u", "myunit"}, []string{}, "myapp", "myunit", nil},
		{[]string{"-a", "myapp"}, []string{"myapp", "myunit"}, "", "", fmt.Errorf("specify app and unit either by flags or by arguments, not both")},
		{[]string{"-u", "myunit"}, []string{"myapp"}, "myapp", "myunit", nil},
		{[]string{"-a", "myapp"}, []string{"myapp", "myunit"}, "", "", fmt.Errorf("specify app and unit either by flags or by arguments, not both")},
		{[]string{"-u", "myunit"}, []string{"myapp", "myunit"}, "", "", fmt.Errorf("specify app and unit either by flags or by arguments, not both")},
		{[]string{"-a", "myapp", "-u", "myunit"}, []string{"myunit"}, "", "", fmt.Errorf("specify app and unit either by flags or by arguments, not both")},
		{[]string{"-a", "myapp", "-u", "myunit"}, []string{"myapp", "myunit"}, "", "", fmt.Errorf("specify app and unit either by flags or by arguments, not both")},
		{[]string{}, []string{"myapp", "myunit", "tooMany"}, "", "", fmt.Errorf("too many arguments")},
	} {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			cmd := newCmd(tsuructx.TsuruContextWithConfig(nil))
			cmd.ParseFlags(test.flags)
			app, unit, err := AppNameAndUnitIDFromArgsOrFlags(cmd, test.args)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expectApp, app)
			assert.Equal(t, test.expectUnit, unit)
		})
	}

}
