package initialize

import (
	"bytes"
	"github.com/ghodss/yaml"
	"github.com/solo-io/solo-kit/cmd/cli/options"
	"github.com/solo-io/solo-kit/cmd/cli/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/solo-io/solo-kit/pkg/errors"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "Initialize new solo-kit project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProject(opts)
		},
	}
	return cmd
}

type Resource struct {
	ResourceName string
	PluralName   string
	ShortName    string
}

func InitProject(opts *options.Options) error {
	return initProject(opts)
}

func initProject(opts *options.Options) error {

	err := initSurvey(opts)
	if err != nil {
		return err
	}
	rootPath := opts.Config.ProjectName
	apiPath := filepath.Join(rootPath, "pkg", "api", "v1")
	err = os.MkdirAll(apiPath, 0755)
	if err != nil {
		return err
	}

	err = genSoloKitYamlFile(rootPath, util.SOLO_KIT_YAML, &opts.Config)
	if err != nil {
		return err
	}

	// resource.proto
	for _, resource := range opts.Init.Resources {
		if len(resource) < 3 {
			return errors.Errorf("resource name must be >= 3 characters")
		}
		shortName := strings.ToLower(resource[:2])
		pluralName := strings.ToLower(resource + "s")
		comResBuf, err := genTemplateBuffer(apiPath,
			resource_proto_common,
			opts.Config,
		)
		if err != nil {
			return err
		}
		resBuf, err := genTemplateBuffer(apiPath,
			resource_proto,
			Resource{
				ResourceName: strcase.ToCamel(resource),
				PluralName:   pluralName,
				ShortName:    shortName,
			},
		)
		if err != nil {
			return err
		}
		err = genFile(apiPath,
			strcase.ToLowerCamel(resource)+".proto",
			comResBuf,
			resBuf,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func initSurvey(opts *options.Options) error {
	var err error
	err = basicInfoSurvey(&opts.Config)
	if err != nil {
		return err
	}

	err = resourceSurvey(&opts.Init)
	if err != nil {
		return err
	}

	err = envSurvey(&opts.Config)
	if err != nil {
		return err
	}
	return nil
}

func genTemplateBuffer(filename, contents string, data interface{}) (*bytes.Buffer, error) {
	tmpl, err := template.New(filename).Parse(contents)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func genFile(path, filename string, contents ...*bytes.Buffer) error {
	var byteArr []byte
	for _, v := range contents {
		byteArr = append(byteArr, v.Bytes()...)
	}
	err := ioutil.WriteFile(filepath.Join(path, filename), byteArr, 0644)
	if err != nil {
		return err
	}
	return nil
}

func genYamlFile(path, filename string, data interface{}) error {
	cfgyml, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(path, filename), cfgyml, 0644)
	if err != nil {
		return err
	}
	return nil
}

func genSoloKitYamlFile(path, filename string, cfg *options.Config) error {
	buf, err := genTemplateBuffer(filename, generate_yaml, *cfg)
	if err != nil {
		return err
	}
	err = genFile(path, filename, buf)
	if err != nil {
		return err
	}
	return nil
}
