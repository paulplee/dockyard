// Package main is the dockyard CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/paulplee/dockyard/cli"
	"github.com/paulplee/dockyard/pkg/dockyard"
)

// version is populated by -ldflags at build time.
var version = "dev"

func main() {
	engine := dockyard.DefaultEngine()
	if err := cli.NewRootCmd(engine, version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}