package codegen

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/solo-io/solo-kit/pkg/errors"

	"github.com/iancoleman/strcase"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func GenerateResourceGroupFiles(apiGroup *model.ApiGroup, skipOutOfPackageFiles, skipGeneratedTests bool) (code_generator.Files, error) {
	var files code_generator.Files

	for _, grp := range apiGroup.ResourceGroupsFoo {
		// TODO joekelley this check probably doesn't make sense
		if skipOutOfPackageFiles && !(strings.HasSuffix(grp.Name, "."+apiGroup.Name) || grp.Name == apiGroup.Name) {
			continue
		}
		fs, err := generateFilesForResourceGroup(grp)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}

	for i := range files {
		files[i].Content = fileHeader + files[i].Content
	}
	if skipGeneratedTests {
		var filesWithoutTests code_generator.Files
		for _, file := range files {
			if strings.HasSuffix(file.Filename, "_test.go") {
				continue
			}
			filesWithoutTests = append(filesWithoutTests, file)
		}
		files = filesWithoutTests
	}
	return files, nil
}

func generateFilesForResourceGroup(rg *model.ResourceGroup) (code_generator.Files, error) {
	var v code_generator.Files
	for suffix, tmpl := range map[string]*template.Template{
		"_snapshot.sk.go":                templates.ResourceGroupSnapshotTemplate,
		"_snapshot_simple_emitter.sk.go": templates.SimpleEmitterTemplate,
		"_snapshot_emitter.sk.go":        templates.ResourceGroupEmitterTemplate,
		"_snapshot_emitter_test.go":      templates.ResourceGroupEmitterTestTemplate,
		"_event_loop.sk.go":              templates.ResourceGroupEventLoopTemplate,
		"_simple_event_loop.sk.go":       templates.SimpleEventLoopTemplate,
		"_event_loop_test.go":            templates.ResourceGroupEventLoopTestTemplate,
	} {
		content, err := generateResourceGroupFile(rg, tmpl)
		if err != nil {
			return nil, errors.Wrapf(err, "internal error: processing %template '%v' for resource group %v failed", tmpl.ParseName, rg.Name)
		}
		v = append(v, code_generator.File{
			Filename: strcase.ToSnake(rg.GoName) + suffix,
			Content:  content,
		})
	}
	return v, nil
}

func generateResourceGroupFile(rg *model.ResourceGroup, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, rg); err != nil {
		return "", err
	}
	return buf.String(), nil
}
