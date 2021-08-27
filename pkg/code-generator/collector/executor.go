package collector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/solo-io/solo-kit/pkg/errors"
)

type ProtocExecutor interface {
	Execute(protoFile string, toFile string, imports []string) error
}

type DefaultProtocExecutor struct {
	// The output directory
	OutputDir string
	// whether or not to do a regular go-proto generate while collecting descriptors
	ShouldCompileFile func(string) bool
	// arguments for go_out=
	CustomGoArgs []string
	// custom plugins
	// each will append a <plugin>_out= directive to protoc command
	CustomPlugins []string
}

var defaultGoArgs = []string{
	"plugins=grpc",
}

func (d *DefaultProtocExecutor) Execute(protoFile string, toFile string, imports []string) error {
	cmd := exec.Command("protoc")

	for _, i := range imports {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-I%s", i))
	}

	if d.ShouldCompileFile(protoFile) {
		goArgs := append(defaultGoArgs, d.CustomGoArgs...)
		goArgsJoined := strings.Join(goArgs, ",")
		cmd.Args = append(cmd.Args,
			fmt.Sprintf("--go_out=%s:%s", goArgsJoined, d.OutputDir),
			fmt.Sprintf("--ext_out=%s:%s", goArgsJoined, d.OutputDir),
		)

		for _, pluginName := range d.CustomPlugins {
			cmd.Args = append(cmd.Args,
				fmt.Sprintf("--%s_out=%s:%s", pluginName, goArgsJoined, d.OutputDir),
			)
		}
	}

	cmd.Args = append(cmd.Args,
		"-o",
		toFile,
		"--include_imports",
		"--include_source_info",
		protoFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, out)
	}
	return nil
}

type OpenApiProtocExecutor struct {
	OutputDir string

	// Whether to include descriptions in validation schemas
	IncludeDescriptionsInSchema bool

	// The maximum number of characters to include in a description
	// A 0 value will be interpreted as "include all characters"
	// Default: 0
	MaxDescriptionCharacters int

	// Whether to assign Enum fields the `x-kubernetes-int-or-string` property
	// which allows the value to either be an integer or a string
	EnumAsIntOrString bool
}

func (o *OpenApiProtocExecutor) Execute(protoFile string, toFile string, imports []string) error {
	cmd := exec.Command("protoc")

	for _, i := range imports {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-I%s", i))
	}

	// The way that --openapi_out works, is that it produces a file in an output directory,
	// with the name of the file matching the proto package (ie gloo.solo.io).
	// Therefore, if you have multiple protos in a single package, they will all be output
	// to the same file, and overwrite one another.
	// To avoid this, we generate a directory with the name of the proto file.
	// For example my_resource.proto in the gloo.solo.io package will produce the following file:
	//  my_resource/gloo.solo.io.yaml

	// The directoryName is created by taking the name of the file and removing the extension
	_, fileName := filepath.Split(protoFile)
	directoryName := fileName[0 : len(fileName)-len(filepath.Ext(fileName))]

	// Create the directory
	directoryPath := filepath.Join(o.OutputDir, directoryName)
	_ = os.Mkdir(directoryPath, os.ModePerm)

	cmd.Args = append(cmd.Args,
		fmt.Sprintf("--openapi_out=yaml=true,single_file=false,max_description_characters=%d,include_description=%v,enum_as_int_or_string=%v:%s",
			o.MaxDescriptionCharacters, o.IncludeDescriptionsInSchema, o.EnumAsIntOrString, directoryPath),
	)

	cmd.Args = append(cmd.Args,
		"-o",
		toFile,
		"--include_imports",
		"--include_source_info",
		protoFile)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%v failed: %s", cmd.Args, out)
	}
	return nil
}
