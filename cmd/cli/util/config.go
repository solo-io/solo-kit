package util

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/solo-io/solo-kit/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	SOLO_KIT_YAML = "solo-kit.yaml"
)

func EnsureConfigFile(opts *options.Options) error {
	if opts.ConfigFile == "" {
		// Check the current file tree

		searchPath := os.ExpandEnv("$GOPATH")
		err := filepath.Walk(searchPath, func(dir string, info os.FileInfo, err error) error {
			if info.IsDir() {
				files, err := ioutil.ReadDir(dir)
				if err != nil {
					return err
				}
				for _, v := range files {
					if v.Name() == SOLO_KIT_YAML {
						configFilePath := path.Join(dir, v.Name())
						opts.ConfigFile = configFilePath
						// Set root as config file path to begin with, in case none is supplied
						opts.Config.Root = configFilePath
					}
				}
			}
			return nil
		})
		if err != nil || opts.ConfigFile == "" {
			return errors.Errorf("Unable to find config file in PATH")
		}

	}
	return nil
}


func ReadConfigFile(opts *options.Options) error {
	data, err := ioutil.ReadFile(opts.ConfigFile)
	if err != nil {
		return errors.Errorf("error reading file: %v", err)
	}
	jsn, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsn, &opts.Config)
}
