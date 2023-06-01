// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"io/fs"
	"path/filepath"

	"github.com/tsuru/tsuru-client/internal/config"
	"github.com/tsuru/tsuru-client/internal/tsuructx"
)

func findExecutablePlugin(tsuruCtx *tsuructx.TsuruContext, pluginName string) (execPath string) {
	basePath := filepath.Join(config.ConfigPath, "plugins")
	testPathGlobs := []string{
		filepath.Join(basePath, pluginName),
		filepath.Join(basePath, pluginName, pluginName),
		filepath.Join(basePath, pluginName, pluginName+".*"),
		filepath.Join(basePath, pluginName+".*"),
	}
	for _, pathGlob := range testPathGlobs {
		var fStat fs.FileInfo
		var err error
		execPath = pathGlob
		if fStat, err = tsuruCtx.Fs.Stat(pathGlob); err != nil {
			files, _ := filepath.Glob(pathGlob)
			if len(files) != 1 {
				continue
			}
			execPath = files[0]
			fStat, err = tsuruCtx.Fs.Stat(execPath)
		}
		if err != nil || fStat.IsDir() || !fStat.Mode().IsRegular() {
			continue
		}
		return execPath
	}
	return ""
}
