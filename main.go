// Copyright © 2016 Gus Esquivel <gesquive@gmail.com>

package main

import "os"
import "fmt"
import "path/filepath"
import "github.com/gesquive/fast-cli/cmd"

var version = "0.1.0"
var dirty = ""
var displayVersion string

func main() {
	displayVersion = fmt.Sprintf("%s v%s%s",
		filepath.Base(os.Args[0]),
		version,
		dirty)
	cmd.Execute(displayVersion)
}
