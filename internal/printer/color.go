// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printer

import (
	"fmt"
)

const (
	pattern  = "\033[%d;%d;%dm%s\033[0m"
	bgFactor = 10
)

var (
	fontColors = map[string]int{
		"black":   30,
		"red":     31,
		"green":   32,
		"yellow":  33,
		"blue":    34,
		"magenta": 35,
		"cyan":    36,
		"white":   37,
	}

	fontEffects = map[string]int{
		"reset":   0,
		"bold":    1,
		"inverse": 7,
	}
)

type Colorify struct {
	DisableColors bool
}

func (c Colorify) Colorfy(msg string, fontcolor string, background string, effect string) string {
	if c.DisableColors {
		return msg
	}
	return fmt.Sprintf(pattern, fontEffects[effect], fontColors[fontcolor], fontColors[background]+bgFactor, msg)
}
