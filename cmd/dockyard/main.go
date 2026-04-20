// Package main is the dockyard CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/paulplee/dockyard/internal/cli"
)

// version is populated by -ldflags at build time.
var version = "dev"

func main() {
	if err := cli.NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
