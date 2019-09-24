package docgen

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/solo-io/go-utils/errors"
	"gopkg.in/yaml.v2"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/datafile"

	"github.com/iancoleman/strcase"
	"github.com/ilackarms/protokit"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	md "github.com/solo-io/solo-kit/pkg/code-generator/docgen/templates/markdown"
	rst "github.com/solo-io/solo-kit/pkg/code-generator/docgen/templates/restructured"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

type DocsGen struct {
	DocsOptions options.DocsOptions
	Project     *model.Project
}

// must ignore validate.proto from lyft
// may need to add more here
var ignoredFiles = []string{
	"validate/validate.proto",
	"github.com/solo-io/solo-kit/api/external/validate/validate.proto",
}

var (
	GenDataFileError = func(err error) error {
		return errors.Wrapf(err, "unable to generate data file")
	}
)

func (d *DocsGen) protoSuffix() string {
	if d.DocsOptions.Output == options.Restructured {
		return ".sk.rst"
	}
	return ".sk.md"
}

func (d *DocsGen) protoFileTemplate() *template.Template {
	if d.DocsOptions.Output == options.Restructured {
		return rst.ProtoFileTemplate(d.Project, &d.DocsOptions)
	}
	return md.ProtoFileTemplate(d.Project, &d.DocsOptions)
}

func (d *DocsGen) GenerateFilesForProtoFiles(protoFiles []*protokit.FileDescriptor) (code_generator.Files, error) {

	// Collect names of files that contain resources for which doc gen has to be skipped
	skipMap := make(map[string]bool)
	for _, res := range d.Project.Resources {
		if res.SkipDocsGen {
			skipMap[res.Filename] = true
		}
	}

	var v code_generator.Files
	for suffix, tmpl := range map[string]*template.Template{
		d.protoSuffix(): d.protoFileTemplate(),
	} {
		for _, protoFile := range protoFiles {
			// Skip if file is to be ignored
			var ignore bool
			for _, ignoredFile := range ignoredFiles {
				if protoFile.GetName() == ignoredFile {
					ignore = true
					break
				}
			}
			if ignore {
				continue
			}

			// Skip if the file contains a top-level resource that has to be skipped
			if skipMap[protoFile.GetName()] {
				continue
			}

			content, err := generateProtoFileFile(protoFile, tmpl)
			if err != nil {
				return nil, err
			}
			fileName := protoFile.GetName() + suffix
			v = append(v, code_generator.File{
				Filename: fileName,
				Content:  content,
			})
		}
	}

	dataFiles, err := d.generateDataFilesForProtoFiles(protoFiles)
	if err != nil {
		return nil, GenDataFileError(err)
	}
	for _, file := range dataFiles {
		v = append(v, file)
	}

	return v, nil
}

func (d *DocsGen) generateDataFilesForProtoFiles(protoFiles []*protokit.FileDescriptor) (code_generator.Files, error) {
	var dataFiles code_generator.Files
	switch d.DocsOptions.Output {
	case options.Hugo:
		return d.generateHugoDataFiles(protoFiles)
	default:
		return dataFiles, nil
	}
}

func (d *DocsGen) generateHugoDataFiles(protoFiles []*protokit.FileDescriptor) (code_generator.Files, error) {
	var dataFiles code_generator.Files
	hugoPbData := datafile.HugoProtobufData{
		Apis: make(map[string]datafile.ApiSummary),
	}
	df := code_generator.File{}
	dfBytes, err := yaml.Marshal(hugoPbData)
	if err != nil {
		return nil, err
	}
	df.Content = string(dfBytes)
	df.Filename = datafile.HugoProtobufRelativeDataPath

	return dataFiles, nil
}

func GenerateFiles(project *model.Project, docsOptions *options.DocsOptions) (code_generator.Files, error) {
	protoFiles := protokit.ParseCodeGenRequest(project.Request)
	if docsOptions == nil {
		docsOptions = &options.DocsOptions{}
	}

	if docsOptions.Output == "" {
		docsOptions.Output = options.Markdown
	}

	docGenerator := DocsGen{
		DocsOptions: *docsOptions,
		Project:     project,
	}

	files, err := docGenerator.GenerateFilesForProject()
	if err != nil {
		return nil, err
	}
	messageFiles, err := docGenerator.GenerateFilesForProtoFiles(protoFiles)
	if err != nil {
		return nil, err
	}
	files = append(files, messageFiles...)

	for i := range files {
		files[i].Content = docGenerator.FileHeader(files[i].Filename) + files[i].Content
	}
	return files, nil
}

func (d *DocsGen) FileHeader(filename string) string {
	if d.DocsOptions.Output == options.Restructured {
		return ".. Code generated by solo-kit. DO NOT EDIT."
	}
	if d.DocsOptions.Output == options.Hugo {
		name := filepath.Base(filename)
		if strings.HasSuffix(name, d.protoSuffix()) {
			name = name[:len(name)-len(d.protoSuffix())]
		}
		return fmt.Sprintf(`
---
title: "%s"
weight: 5
---

<!-- Code generated by solo-kit. DO NOT EDIT. -->

`, name)
	}
	return `<!-- Code generated by solo-kit. DO NOT EDIT. -->
`
}

func (d *DocsGen) projectSuffix() string {
	if d.DocsOptions.Output == options.Restructured {
		return ".project.sk.rst"
	}
	return ".project.sk.md"
}

func (d *DocsGen) projectDocsRootTemplate() *template.Template {
	if d.DocsOptions.Output == options.Restructured {
		return rst.ProjectDocsRootTemplate(d.Project, &d.DocsOptions)
	}
	return md.ProjectDocsRootTemplate(d.Project, &d.DocsOptions)
}

func (d *DocsGen) GenerateFilesForProject() (code_generator.Files, error) {
	var v code_generator.Files
	for suffix, tmpl := range map[string]*template.Template{
		d.projectSuffix(): d.projectDocsRootTemplate(),
	} {
		content, err := generateProjectFile(d.Project, tmpl)
		if err != nil {
			return nil, err
		}
		v = append(v, code_generator.File{
			Filename: strcase.ToSnake(d.Project.ProjectConfig.Name) + suffix,
			Content:  content,
		})
	}
	return v, nil
}

func generateProjectFile(project *model.Project, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, project); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generateProtoFileFile(protoFile *protokit.FileDescriptor, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, protoFile); err != nil {
		return "", err
	}
	return buf.String(), nil
}
