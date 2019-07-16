package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/ilackarms/protokit"
	"github.com/solo-io/solo-kit/pkg/errors"
)

const ProjectConfigFilename = "solo-kit.json"

// SOLO-KIT Descriptors from which code can be generated

type SoloKitProject struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ApiGroups   []*ApiGroup `json:"api_groups"`

	// set by load
	ProjectFile string
}

type ApiGroup struct {
	Name                   string                      `json:"name"`
	DocsDir                string                      `json:"docs_dir"`
	VersionConfigs         []*VersionConfig            `json:"version_configs"`
	ResourceGroups         map[string][]ResourceConfig `json:"resource_groups"`
	ResourceGroupGoPackage string                      `json:"resource_group_go_package"`
	ConversionGoPackage    string                      `json:"conversion_go_package"`

	// if set, this group will override the proto package typically used
	// as the api group for the crd
	CrdGroupOverride string `json:"crd_group_override"`

	// imported solokit projects, used for resource groups
	Imports []string `json:"imports"`

	// set by load
	SoloKitProject *SoloKitProject
	Conversions    []*Conversion
	// TODO joekelley improve name
	ResourceGroupsFoo           []*ResourceGroup
	ConversionGoPackageShort    string
	ResourceGroupGoPackageShort string
}

func (a ApiGroup) IsOurProto(protoFile string) bool {
	for _, vc := range a.VersionConfigs {
		if vc.IsOurProto(protoFile) {
			return true
		}
	}
	return false
}

type VersionConfig struct {
	Version string `json:"version"`

	// define custom resources here
	CustomResources []CustomResourceConfig `json:"custom_resources"`
	// set by load if not specified
	GoPackage string `json:"go_package"`

	// set by load
	ApiGroup      *ApiGroup
	VersionProtos []string
}

func (p VersionConfig) IsOurProto(protoFile string) bool {
	for _, file := range p.VersionProtos {
		if protoFile == file {
			return true
		}
	}
	return false
}

type ResourceConfig struct {
	ResourceName    string `json:"name"`
	ResourcePackage string `json:"package"` // resource package doubles as the proto package or the go import package
}

// Create a Solo-Kit backed resource from
// a Go Type that implements the Resource Interface
type CustomResourceConfig struct {
	// the import path for the Go Type
	Package string `json:"package"`
	// the name of the Go Type
	Type          string `json:"type"`
	PluralName    string `json:"plural_name"`
	ShortName     string `json:"short_name"`
	ClusterScoped bool   `json:"cluster_scoped"`

	// set by load
	Imported bool
}

type Version struct {
	VersionConfig VersionConfig
	ProtoPackage  string
	Resources     []*Resource
	XDSResources  []*XDSResource

	Request *plugin_go.CodeGeneratorRequest
}

type Resource struct {
	Name         string
	PluralName   string
	ShortName    string
	ProtoPackage string // eg. gloo.solo.io
	// ImportPrefix will equal ProtoPackage+"." if the resource does not belong to the project
	// else it will be empty string. used in event loop files
	ImportPrefix string
	// empty unless resource is external
	// format "github.com/solo-io/solo-kit/foo/bar"
	GoPackage string

	HasStatus          bool
	ClusterScoped      bool                 // the resource lives at the cluster level, namespace is ignored
	SkipDocsGen        bool                 // if true, no docs will be generated for the proto file where this resource is defined
	IsCustom           bool                 // if true, this will be treated as a custom resource without a proto file behind it
	CustomResource     CustomResourceConfig // this struct will be empty unless IsCustom is true
	CustomImportPrefix string               // import prefix for the struct type the generated wrapper will wrap

	Fields []*Field
	Oneofs []*Oneof

	// resource groups i belong to
	ResourceGroups []*ResourceGroup
	// project i belong to
	Project *Version

	Filename string // the proto file where this resource is contained
	Version  string // set during parsing from this resource's solo-kit.json

	// TODO (ilackarms): change to use descriptor.DescriptorProto
	Original *protokit.Descriptor
}

type Field struct {
	Name        string
	TypeName    string
	IsOneof     bool // we ignore oneof fields in test generation
	SkipHashing bool // skip this field when hashing the resource
	Original    *descriptor.FieldDescriptorProto
}

type Oneof struct {
	Name string
}

type ResourceGroup struct {
	Name      string // eg. api.gloo.solo.io
	GoName    string // will be Api
	Imports   string // if this resource group contains any imports from other projects
	ApiGroup  *ApiGroup
	Resources []*Resource
}

type XDSResource struct {
	Name         string
	MessageType  string
	NameField    string
	NoReferences bool

	Project      *Version
	ProtoPackage string // eg. gloo.solo.io

	Filename string // the proto file where this resource is contained
}

func LoadProjectConfig(path string) (SoloKitProject, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return SoloKitProject{}, err
	}
	var skp SoloKitProject
	err = json.Unmarshal(b, &skp)
	if err != nil {
		return SoloKitProject{}, err
	}

	skp.ProjectFile = path
	for _, ag := range skp.ApiGroups {
		goPackageSegments := strings.Split(ag.ResourceGroupGoPackage, "/")
		ag.ResourceGroupGoPackageShort = goPackageSegments[len(goPackageSegments)-1]
		for _, vc := range ag.VersionConfigs {
			if vc.GoPackage == "" {
				goPkg, err := detectGoPackageForVersion(filepath.Dir(skp.ProjectFile) + "/" + vc.Version)
				if err != nil {
					return SoloKitProject{}, err
				}
				vc.GoPackage = goPkg
			}
		}
	}

	return skp, err
}

var goPackageStatementRegex = regexp.MustCompile(`option go_package.*=.*"(.*)";`)

// Returns the value of the 'go_package' option of the first .proto file found in the version's directory
func detectGoPackageForVersion(versionDir string) (string, error) {
	var goPkg string
	if err := filepath.Walk(versionDir, func(protoFile string, info os.FileInfo, err error) error {
		// already set
		if goPkg != "" {
			return nil
		}
		if !strings.HasSuffix(protoFile, ".proto") {
			return nil
		}
		// search for go_package on protos in the same dir as the project.json
		if versionDir != filepath.Dir(protoFile) {
			return nil
		}
		content, err := ioutil.ReadFile(protoFile)
		if err != nil {
			return err
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			goPackage := goPackageStatementRegex.FindStringSubmatch(line)
			if len(goPackage) == 0 {
				continue
			}
			if len(goPackage) != 2 {
				return errors.Errorf("parsing go_package error: from %v found %v", line, goPackage)
			}
			goPkg = goPackage[1]
			break
		}
		return nil
	}); err != nil {
		return "", err
	}
	if goPkg == "" {
		return "", errors.Errorf("no go_package statement found in root dir of project %v", versionDir)
	}
	return goPkg, nil
}
