package collector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/solo-io/solo-kit/pkg/code-generator/model"
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
	// each will append a <plugin_key>_out=<plugin_value> directive to protoc command
	CustomPlugins map[string]string
}

var defaultGoArgs = []string{
	"plugins=grpc",
	"Mgithub.com/solo-io/solo-kit/api/external/envoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2",
}

func (d *DefaultProtocExecutor) Execute(protoFile string, toFile string, imports []string) error {
	cmd := exec.Command("protoc")

	for _, i := range imports {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-I%s", i))
	}

	if d.ShouldCompileFile(protoFile) {
		goArgs := append(defaultGoArgs, d.CustomGoArgs...)
		cmd.Args = append(cmd.Args,
			fmt.Sprintf("--go_out=%s:%s", strings.Join(goArgs, ","), d.OutputDir),
			fmt.Sprintf("--ext_out=%s:%s", strings.Join(goArgs, ","), d.OutputDir),
		)

		for pluginName, pluginOpts := range d.CustomPlugins {
			cmd.Args = append(cmd.Args,
				fmt.Sprintf("--%s_out=%s:%s", pluginName, pluginOpts, d.OutputDir),
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
	Project   *model.Project
}

func (o *OpenApiProtocExecutor) Execute(protoFile string, toFile string, imports []string) error {
	cmd := exec.Command("protoc")

	if !o.Project.ProjectConfig.IsOurProto(protoFile) {
		// We only run open api generation for project protos that describe CRDs
		// If it is not a relevant proto, do nothing
		return nil
	}

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
		fmt.Sprintf("--openapi_out=yaml=true,single_file=false:%s", directoryPath),
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
