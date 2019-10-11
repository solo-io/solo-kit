package docgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/solo-io/solo-kit/pkg/code-generator/parser"

	"gopkg.in/yaml.v2"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/datafile"

	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
)

// write docs that are produced from the content of multiple projects
func WriteCrossProjectDocs(
	genDocs *options.DocsOptions,
	absoluteRoot string,
	projectMap parser.ProjectMap,
) error {
	if genDocs == nil {
		return nil
	}
	switch genDocs.Output {
	case options.Hugo:
		return WriteCrossProjectDocsHugo(genDocs, absoluteRoot, projectMap)
	default:
		return nil
	}
}

// Hugo docs require a datafile that is used by shortcodes
func WriteCrossProjectDocsHugo(
	genDocs *options.DocsOptions,
	absoluteRoot string,
	projectMap parser.ProjectMap,
) error {
	// only write the data file if a data dir is specified
	hugoOptions := genDocs.HugoOptions
	if hugoOptions == nil {
		return nil
	}
	if hugoOptions.DataDir == "" {
		return nil
	}
	hugoPbData := datafile.HugoProtobufData{
		Apis: make(map[string]datafile.ApiSummary),
	}

	for _, project := range projectMap {
		// Collect names of files that contain resources for which doc gen has to be skipped
		skipMap := make(map[string]bool)
		for _, res := range project.Resources {
			if res.SkipDocsGen {
				skipMap[res.Filename] = true
			}
		}
		for _, pf := range project.Descriptors {
			filename := pf.GetName()
			if skipMap[filename] {
				continue
			}
			if pf.Package == nil {
				// if there is no package we will not generate doc links for the descriptor
				continue
			}
			protoPkgName := *pf.Package
			// TODO - apply if we decide to emit a singe docs page per proto pkg, rather than per proto file
			// until then, it's not clear which file should be targeted
			// package-level page link
			//key, value := getApiSummaryKV(hugoOptions.ApiDir, filename, *pf.Package, "")
			//hugoPbData.Apis[key] = value
			for _, message := range pf.Messages {
				if message.Name == nil {
					fmt.Printf("underspecified message will not be added to the map: %v", message)
					continue
				}
				protoMsgName := *message.Name
				// message-level sub-page link
				key, value := getApiSummaryKV(hugoOptions.ApiDir, filename, protoPkgName, protoMsgName)
				hugoPbData.Apis[key] = value
			}
		}
	}
	fileBytes, err := yaml.Marshal(hugoPbData)
	if err != nil {
		return err
	}

	file := code_generator.File{
		Filename: options.HugoProtoDataFile,
		Content:  string(fileBytes),
	}

	// note that the data file is saved in the DataDir, not the directory specified by the project spec
	path := filepath.Join(absoluteRoot, hugoOptions.DataDir, file.Filename)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, []byte(file.Content), 0644); err != nil {
		return err
	}
	return nil

}

// util for populating the Hugo ProtoMap to enable allows lookup by
//   somepkg.solo.io or somepkg.solo.io.SomeMessage
func getApiSummaryKV(apiDir, filename, packageName, fieldName string) (string, datafile.ApiSummary) {
	key := packageName
	hashPath := ""
	if fieldName != "" {
		hashPath = fmt.Sprintf("#%v", fieldName)
		key = fmt.Sprintf("%v.%v", packageName, fieldName)
	}
	filePath := filepath.Join(apiDir, filename+options.HugoResourceExtension)
	relativePath := fmt.Sprintf("%v%v", filePath, hashPath)
	value := datafile.ApiSummary{
		RelativePath: relativePath,
		Package:      packageName,
	}
	return key, value
}
