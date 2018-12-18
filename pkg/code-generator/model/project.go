package model

import (
	"encoding/json"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pseudomuto/protokit"
	"io/ioutil"
)

// SOLO-KIT Descriptors from which code can be generated

type ProjectConfig struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	DocsDir     string `json:"docs_dir"`
	// set by load
	ProjectRoot string
}

type Project struct {
	ProjectConfig
	GroupName string

	Resources      []*Resource
	ResourceGroups []*ResourceGroup

	XDSResources []*XDSResource

	Request *plugin_go.CodeGeneratorRequest
}

type Resource struct {
	Name       string
	PluralName string
	ShortName  string
	GroupName  string // eg. gloo.solo.io
	// ImportPrefix will equal GroupName+"." if the resource does not belong to the project
	// else it will be empty string. used in event loop files
	ImportPrefix string
	// empty unless resource is external
	// format "github.com/solo-io/solo-kit/foo/bar"
	GoPackage string

	HasStatus bool
	Fields    []*Field

	ResourceGroups []*ResourceGroup
	Project        *Project

	Filename string // the proto file where this resource is contained

	// TODO (ilackarms): change to use descriptor.DescriptorProto
	Original *protokit.Descriptor
}

type Field struct {
	Name     string
	TypeName string
	IsOneof  bool // we ignore oneof fields in test generation
	Original *descriptor.FieldDescriptorProto
}

type ResourceGroup struct {
	Name      string // eg. api.gloo.solo.io
	GoName    string // will be Api
	Imports   string // if this resource group contains any imports from other projects
	Project   *Project
	Resources []*Resource
}

type XDSResource struct {
	Name         string
	MessageType  string
	NameField    string
	NoReferences bool

	Project   *Project
	GroupName string // eg. gloo.solo.io
	Package   string // proto package for the message
}

func LoadProjectConfig(path string) (ProjectConfig, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, err
	}
	var pc ProjectConfig
	err = json.Unmarshal(b, &pc)
	pc.ProjectRoot = path
	return pc, err
}
