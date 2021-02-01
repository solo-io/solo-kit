package main

import (
	"fmt"
	"os"

	"github.com/solo-io/go-list-licenses/pkg/license"
)

func main() {
	glooPackages := []string{
		"github.com/solo-io/solo-kit/cmd/cli",
		"github.com/solo-io/solo-kit/cmd/solo-kit-gen",
	}

	// dependencies for this package which are used on mac, and will not be present in linux CI
	macOnlyDependencies := []string{
		"github.com/mitchellh/go-homedir",
		"github.com/containerd/continuity",
		"golang.org/x/sys/unix",
	}

	app := license.Cli(glooPackages, macOnlyDependencies)
	if err := app.Execute(); err != nil {
		fmt.Errorf("unable to run oss compliance check: %v\n", err)
		os.Exit(1)
	}
}
