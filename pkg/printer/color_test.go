// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorfy(t *testing.T) {
	for _, test := range []struct {
		disableColors          bool
		msg, color, bg, effect string
		expected               string
	}{
		{false, "WORDS", "", "", "", "\033[0;0;10mWORDS\033[0m"},
		{false, "WORDS", "black", "yellow", "bold", "\033[1;30;43mWORDS\033[0m"},
		{true, "WORDS", "", "", "", "WORDS"},
		{true, "WORDS", "black", "yellow", "bold", "WORDS"},
	} {
		c := Colorify{test.disableColors}
		assert.Equal(t, test.expected, c.Colorfy(test.msg, test.color, test.bg, test.effect))
	}
}
