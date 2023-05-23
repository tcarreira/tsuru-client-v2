// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func getToken(fsys afero.Fs) (string, error) {
	if token := viper.GetString("token"); token != "" {
		return token, nil
	}

	tokenPaths := []string{filepath.Join(ConfigPath, "token")}
	if targetLabel, err := getTargetLabel(fsys); err == nil {
		tokenPaths = append([]string{filepath.Join(ConfigPath, "token.d", targetLabel)}, tokenPaths...)
	}

	var err error
	for _, tokenPath := range tokenPaths {
		var tkFile afero.File
		if tkFile, err = fsys.Open(tokenPath); err == nil {
			defer tkFile.Close()
			token, err1 := io.ReadAll(tkFile)
			if err1 != nil {
				return "", err1
			}
			tokenStr := strings.TrimSpace(string(token))
			return tokenStr, nil
		}
	}
	if os.IsNotExist(err) {
		return "", nil
	}
	return "", err
}

// GetToken returns the token for the current target,
// as defined in the TSURU_TOKEN environment variable or in the token file.
func GetToken() (string, error) {
	return getToken(afero.NewOsFs())
}

func saveToken(token string, fsys afero.Fs) error {
	tokenPaths := []string{filepath.Join(ConfigPath, "token")}
	targetLabel, err := getTargetLabel(fsys)
	if err == nil {
		err := fsys.MkdirAll(filepath.Join(ConfigPath, "token.d"), 0700)
		if err != nil {
			return err
		}
		tokenPaths = append(tokenPaths, filepath.Join(ConfigPath, "token.d", targetLabel))
	}
	for _, tokenPath := range tokenPaths {
		file, err := fsys.Create(tokenPath)
		if err != nil {
			return err
		}
		defer file.Close()
		n, err := file.WriteString(token)
		if err != nil {
			return err
		}
		if n != len(token) {
			return fmt.Errorf("failed to write token file")
		}
	}
	return nil
}

// SaveToken returns the token for the current target,
// as defined in the TSURU_TOKEN environment variable or in the token file.
func SaveToken(token string) error {
	return saveToken(token, afero.NewOsFs())
}