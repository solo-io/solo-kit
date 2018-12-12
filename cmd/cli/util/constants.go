package util


import (
"fmt"
"github.com/solo-io/solo-kit/cmd/cli/options"
"os"
)

const (
	root_string = "$GOPATH/src"
	gogo_flag_string = "--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"
)

var (
	ROOT       string
	SOLO_KIT   string
	TEST_IN    string
	OUT        string
	DOCS_OUT   string
	PROJECT_IN string
	IMPORTS    string

	GOGO_FLAG     string
	INPUT_PROTOS  string
)

func init() {
	ROOT = os.ExpandEnv(root_string)

	GOGO_FLAG = os.ExpandEnv(gogo_flag_string)
}

var SOLO_KIT_FLAG = func(cfg *options.Config) string {
	return os.ExpandEnv(fmt.Sprintf("--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=%s --solo-kit_opt=${PWD}/project.json,%s/doc/docs/v1", cfg.Output, cfg.Root))

}
