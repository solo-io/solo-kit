package docgen

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rotisserie/eris"
	"github.com/solo-io/solo-kit/pkg/code-generator/writer"

	"github.com/iancoleman/strcase"
	"github.com/pseudomuto/protokit"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	md "github.com/solo-io/solo-kit/pkg/code-generator/docgen/templates/markdown"
	rst "github.com/solo-io/solo-kit/pkg/code-generator/docgen/templates/restructured"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

type DocsGen struct {
	DocsOptions options.DocsOptions
	Project     *model.Project
	DocsDir     string
}

// must ignore validate.proto from lyft
// may need to add more here
var ignoredFiles = []string{
	"validate/validate.proto",
	"github.com/solo-io/solo-kit/api/external/validate/validate.proto",
	"github.com/envoyproxy/protoc-gen-validate/validate/validate.proto",
}

// write docs that are produced from the content of a single project
func WritePerProjectsDocs(project *model.Project, docsOptions *options.DocsOptions, absoluteRoot string) error {
	if project.ProjectConfig.DocsDir != "" && docsOptions != nil {

		if docsOptions.Output == "" {
			docsOptions.Output = options.Markdown
		}

		docGenerator := &DocsGen{
			DocsOptions: *docsOptions,
			Project:     project,
			DocsDir:     filepath.Join(absoluteRoot, project.ProjectConfig.DocsDir),
		}

		return GenerateAndWriteFiles(docGenerator)
	}
	return nil
}

func GenerateAndWriteFiles(docGenerator *DocsGen) error {
	if docGenerator == nil {
		return eris.New("doc generator is nil")
	}

	files, err := docGenerator.GenerateFilesForProject()
	if err != nil {
		return err
	}

	messageFiles, err := docGenerator.GenerateFilesForProtoFiles(docGenerator.Project.Descriptors)
	if err != nil {
		return err
	}
	files = append(files, messageFiles...)

	return docGenerator.WriteFiles(files)
}

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
				Filename:   fileName,
				Content:    content,
				Permission: 0644,
			})
		}
	}

	return v, nil
}

func (d *DocsGen) WriteFiles(files code_generator.Files) error {
	fileWriter := &writer.DefaultFileWriter{
		Root:               d.DocsDir,
		HeaderFromFilename: d.FileHeader,
	}

	return fileWriter.WriteFiles(files)
}

func (d *DocsGen) FileHeader(filename string) string {
	if d.DocsOptions.Output == options.Restructured {
		return ".. Code generated by solo-kit. DO NOT EDIT."
	}
	if d.DocsOptions.Output == options.Hugo {
		return d.hugoFileHeader(filename)
	}
	return `<!-- Code generated by solo-kit. DO NOT EDIT. -->
`
}

func (d *DocsGen) hugoFileHeader(filename string) string {
	name := filepath.Base(filename)

	var title string
        protoExtension := ".proto"+d.protoSuffix()
	if strings.HasSuffix(name, protoExtension) {
		// Remove the "proto.sk.md" extension
		name = name[:len(name)-len(protoExtension)]

		title = strcase.ToCamel(name)
	} else {
		// Not a file generated by a proto file, leave the title to match the name of the file
		title = name
	}

	return fmt.Sprintf(`
---
title: "%s"
weight: 5
---

<!-- Code generated by solo-kit. DO NOT EDIT. -->

`, title)
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
