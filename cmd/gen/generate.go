// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Keep all go:generate commands here:
//go:generate go run . legacy ../../internal/legacy/legacy.go

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gen <command> [<outputfile>]")
		os.Exit(1)
	}

	buf := bytes.NewBuffer(nil)
	switch os.Args[1] {
	case "legacy":
		genTsuruLegacy(buf)
	default:
		fmt.Printf("Unknown command %s\n", os.Args[1])
		os.Exit(1)
	}

	output := os.Stdout
	if len(os.Args) > 2 {
		var err error
		output, err = os.Create(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer output.Close()

		path := os.Args[2]
		if wd, err := os.Getwd(); err != nil && strings.HasSuffix(wd, "/cmd/gen") {
			path = strings.TrimPrefix(path, "../../")
		}
		fmt.Printf("... Writing generated content to %s\n", path)
	}

	fmt.Fprint(output, buf)

}
