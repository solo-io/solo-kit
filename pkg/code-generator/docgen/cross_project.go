package docgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/solo-io/solo-kit/pkg/code-generator/parser"

	"gopkg.in/yaml.v2"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/datafile"

	"github.com/ilackarms/protokit"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func WriteCrossProjectDocs(projectConfigs []*model.ProjectConfig, genDocs *options.DocsOptions, absoluteRoot string, protoDescriptors []*descriptor.FileDescriptorProto) error {

	switch genDocs.Output {
	case options.Hugo:
		return WriteCrossProjectDocsHugo(projectConfigs, genDocs, absoluteRoot, protoDescriptors)
	default:
		return nil
	}
}

// Hugo docs require a datafile that is used by shortcodes
func WriteCrossProjectDocsHugo(projectConfigs []*model.ProjectConfig, genDocs *options.DocsOptions, absoluteRoot string, protoDescriptors []*descriptor.FileDescriptorProto) error {
	// only write the data file if a data dir is specified
	hugoOptions := genDocs.HugoOptions
	if hugoOptions == nil {
		return nil
	}
	hugoDataDir := hugoOptions.DataDir
	if hugoDataDir == "" {
		return nil
	}
	hugoPbData := datafile.HugoProtobufData{
		Apis: make(map[string]datafile.ApiSummary),
	}
	docsDir := ""

	for _, projectConfig := range projectConfigs {
		project, err := parser.ProcessDescriptors(projectConfig, projectConfigs, protoDescriptors)
		if err != nil {
			return err
		}
		// Collect names of files that contain resources for which doc gen has to be skipped
		skipMap := make(map[string]bool)
		for _, res := range project.Resources {
			if res.SkipDocsGen {
				skipMap[res.Filename] = true
			}
		}
		docsDir = project.ProjectConfig.DocsDir
		protoFiles := protokit.ParseCodeGenRequest(project.Request)
		for _, pf := range protoFiles {
			filename := pf.GetName()
			if skipMap[filename] {
				continue
			}
			// TODO - apply if we decide to emit a singe docs page per proto pkg, rather than per proto file
			// until then, it's not clear which file should be targeted
			// package-level page link
			//key, value := getApiSummaryKV(filename, *pf.Package, "")
			//hugoPbData.Apis[key] = value
			for _, message := range pf.Messages {
				// message-level sub-page link
				key, value := getApiSummaryKV(filename, *pf.Package, *message.Name)
				hugoPbData.Apis[key] = value
			}
		}
	}
	fileBytes, err := yaml.Marshal(hugoPbData)
	if err != nil {
		return err
	}

	file := code_generator.File{
		Filename: filepath.Join(hugoDataDir, options.HugoProtoDataFile),
		Content:  string(fileBytes),
	}

	path := filepath.Join(absoluteRoot, docsDir, file.Filename)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
		return err
	}
	return nil

}

// util for populating the Hugo ProtoMap to enable allows lookup by
//   somepkg.solo.io or SomeMessage.somepkg.solo.io
func getApiSummaryKV(filename, packageName, fieldName string) (string, datafile.ApiSummary) {
	key := packageName
	hashPath := ""
	if fieldName != "" {
		hashPath = fmt.Sprintf("#%v", fieldName)
		key = fmt.Sprintf("%v.%v", packageName, fieldName)
	}
	relativePath := fmt.Sprintf("%v%v", filename, hashPath)
	value := datafile.ApiSummary{
		RelativePath: relativePath,
		Package:      packageName,
	}
	return key, value
}
