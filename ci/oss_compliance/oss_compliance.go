package main

import (
	"fmt"
	"os"

	"github.com/solo-io/go-list-licenses/pkg/license"
)

func main() {
	// dependencies for this package which are used on mac, and will not be present in linux CI
	macOnlyDependencies := []string{
		"github.com/mitchellh/go-homedir",
		"github.com/containerd/continuity",
		"golang.org/x/sys/unix",
		"golang.org/x/sys",
	}

	app, err := license.CliAllPackages(macOnlyDependencies)
	if err != nil {
		fmt.Printf("unable to list all solo-kit dependencies: %v\n", err)
		os.Exit(1)
	}
	if err := app.Execute(); err != nil {
		fmt.Printf("unable to run oss compliance check: %v\n", err)
		os.Exit(1)
	}
}
