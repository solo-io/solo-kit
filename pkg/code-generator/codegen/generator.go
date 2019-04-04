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
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

const fileHeader = `// Code generated by solo-kit. DO NOT EDIT.

`

func GenerateFiles(project *model.Project, skipOutOfPackageFiles bool) (code_generator.Files, error) {
	files, err := generateFilesForProject(project)
	if err != nil {
		return nil, err
	}
	for _, res := range project.Resources {
		// only generate files for the resources in our group, otherwise we import
		if res.ProtoPackage != project.ProtoPackage && !res.IsCustom {
			log.Printf("not generating solo-kit "+
				"clients for resource %v.%v, "+
				"resource proto package must match project proto package %v", res.ProtoPackage, res.Name, project.ProtoPackage)
			continue
		}
		fs, err := generateFilesForResource(res)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	for _, grp := range project.ResourceGroups {
		if skipOutOfPackageFiles && !(strings.HasSuffix(grp.Name, "."+project.ProtoPackage) || grp.Name == project.ProtoPackage) {
			continue
		}
		fs, err := generateFilesForResourceGroup(grp)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}

	for _, res := range project.XDSResources {
		if skipOutOfPackageFiles && res.ProtoPackage != project.ProtoPackage && !strings.HasSuffix(res.ProtoPackage, "."+project.ProtoPackage) {
			continue
		}
		fs, err := generateFilesForXdsResource(res)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}
	for i := range files {
		files[i].Content = fileHeader + files[i].Content
	}
	return files, nil
}

func generateFilesForXdsResource(resource *model.XDSResource) (code_generator.Files, error) {
	var v code_generator.Files
	for suffix, tmpl := range map[string]*template.Template{
		"_xds.sk.sk.go": templates.XdsTemplate,
	} {
		content, err := generateXdsResourceFile(resource, tmpl)
		if err != nil {
			return nil, err
		}
		v = append(v, code_generator.File{
			Filename: strcase.ToSnake(resource.Name) + suffix,
			Content:  content,
		})
	}
	return v, nil
}

func generateFilesForResource(resource *model.Resource) (code_generator.Files, error) {
	var v code_generator.Files
	for suffix, tmpl := range map[string]*template.Template{
		".sk.go":            templates.ResourceTemplate,
		"_client.sk.go":     templates.ResourceClientTemplate,
		"_client_test.go":   templates.ResourceClientTestTemplate,
		"_reconciler.sk.go": templates.ResourceReconcilerTemplate,
	} {
		content, err := generateResourceFile(resource, tmpl)
		if err != nil {
			return nil, errors.Wrapf(err, "internal error: processing template '%v' for resource %v failed", tmpl.ParseName, resource.Name)
		}
		v = append(v, code_generator.File{
			Filename: strcase.ToSnake(resource.Name) + suffix,
			Content:  content,
		})
	}
	return v, nil
}

func generateFilesForResourceGroup(rg *model.ResourceGroup) (code_generator.Files, error) {
	var v code_generator.Files
	for suffix, tmpl := range map[string]*template.Template{
		"_snapshot.sk.go":           templates.ResourceGroupSnapshotTemplate,
		"_snapshot_emitter.sk.go":   templates.ResourceGroupEmitterTemplate,
		"_snapshot_emitter_test.go": templates.ResourceGroupEmitterTestTemplate,
		"_event_loop.sk.go":         templates.ResourceGroupEventLoopTemplate,
		"_event_loop_test.go":       templates.ResourceGroupEventLoopTestTemplate,
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

func generateFilesForProject(project *model.Project) (code_generator.Files, error) {
	var v code_generator.Files
	for suffix, tmpl := range map[string]*template.Template{
		"_suite_test.go": templates.ProjectTestSuiteTemplate,
	} {
		content, err := generateProjectFile(project, tmpl)
		if err != nil {
			return nil, errors.Wrapf(err, "internal error: processing template '%v' for project %v failed", tmpl.ParseName, project.ProjectConfig.Name)
		}
		v = append(v, code_generator.File{
			Filename: strcase.ToSnake(project.ProjectConfig.Name) + suffix,
			Content:  content,
		})
	}
	return v, nil
}

func generateXdsResourceFile(resource *model.XDSResource, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, resource); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generateResourceFile(resource *model.Resource, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, resource); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generateResourceGroupFile(rg *model.ResourceGroup, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, rg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generateProjectFile(project *model.Project, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, project); err != nil {
		return "", err
	}
	return buf.String(), nil
}
