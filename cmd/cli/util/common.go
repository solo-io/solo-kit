package util

import (
	"fmt"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	SOLO_KIT_YAML = "solo-kit.yaml"
)

func EnsureConfig(cfg *options.Config) {
	if cfg.Dir == "" {
		// Check the current file tree

		searchPath := (os.ExpandEnv("$GOPATH"))
		err := filepath.Walk(searchPath, func(dir string, info os.FileInfo, err error) error {
			if info.IsDir() {
				files, err := ioutil.ReadDir(dir)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				for _, v := range files {
					if v.Name() == SOLO_KIT_YAML {
						cfg.Dir = path.Join(dir, v.Name())
					}
				}
			}
			return nil
		})
		if err != nil || cfg.Dir == "" {
			fmt.Println("Unable to find config file in PATH")
			os.Exit(1)
		}

	}

	// Use config file from the flag.
	viper.SetConfigFile(cfg.Dir)

	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
