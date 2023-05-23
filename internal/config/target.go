// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var (
	errUndefinedTarget = fmt.Errorf(`no target defined. Please use target-add/target-set to define a target.

For more details, please run "tsuru help target"`)
)

// getTargets returns a map of label->target
func getTargets(fsys afero.Fs) (map[string]string, error) {
	var targets = map[string]string{} // label->target

	// legacyTargetsPath := JoinWithUserDir(".tsuru_targets") // XXX: remove legacy file
	targetsPath := filepath.Join(ConfigPath, "targets")
	err := fsys.MkdirAll(ConfigPath, 0700)
	if err != nil {
		return nil, err
	}

	f, err := fsys.Open(targetsPath)
	if err == nil {
		defer f.Close()
		if b, err := io.ReadAll(f); err == nil {
			var targetLines = strings.Split(strings.TrimSpace(string(b)), "\n")
			for i := range targetLines {
				var targetSplit = strings.Fields(targetLines[i])

				if len(targetSplit) == 2 {
					targets[targetSplit[0]] = targetSplit[1]
				}
			}
		}
	}
	return targets, nil
}

func getTargetLabel(fsys afero.Fs) (string, error) {
	target, err := getTarget(fsys)
	if err != nil {
		return "", err
	}
	targets, err := getTargets(fsys)
	if err != nil {
		return "", err
	}
	targetKeys := make([]string, len(targets))
	for k := range targets {
		targetKeys = append(targetKeys, k)
	}
	sort.Strings(targetKeys)
	for _, k := range targetKeys {
		if targets[k] == target {
			return k, nil
		}
	}
	return "", fmt.Errorf("label for target %q not found ", target)

}

// getTarget returns the current target,
// as defined in the TSURU_TARGET environment variable or in the target file.
func getTarget(fsys afero.Fs) (target string, err error) {
	if target = viper.GetString("target"); target != "" {
		targets, err := getTargets(fsys)
		if err == nil {
			if val, ok := targets[target]; ok {
				target = val
			}
		}
	} else {
		targetPath := filepath.Join(ConfigPath, "target")
		if f, err := fsys.Open(targetPath); err == nil {
			defer f.Close()
			if b, err := io.ReadAll(f); err == nil {
				target = strings.TrimSpace(string(b))
			}
		}
	}

	if target == "" {
		return "", errUndefinedTarget
	}

	var prefix string
	if m, _ := regexp.MatchString("^https?://", target); !m {
		prefix = "http://"
	}
	return prefix + target, nil
}

// GetTarget returns the current target,
// as defined in the TSURU_TARGET environment variable or in the target file.
func GetTarget() (string, error) {
	return getTarget(afero.NewOsFs())
}
