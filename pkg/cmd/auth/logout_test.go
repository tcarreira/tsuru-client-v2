// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogoutCmd(t *testing.T) {
	assert.NotNil(t, NewLogoutCmd())
}
