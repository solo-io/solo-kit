package codegen

import (
	"bytes"
	"sort"
	"text/template"

	"github.com/solo-io/go-utils/versionutils/kubeapi"
	"github.com/solo-io/solo-kit/pkg/errors"

	"github.com/solo-io/go-utils/log"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func GenerateConversionFiles(soloKitProject *model.ApiGroup, versions []*model.Version) (code_generator.Files, error) {
	var files code_generator.Files

	sort.SliceStable(versions, func(i, j int) bool {
		vi, err := kubeapi.ParseVersion(versions[i].VersionConfig.Version)
		if err != nil {
			return false
		}
		vj, err := kubeapi.ParseVersion(versions[j].VersionConfig.Version)
		if err != nil {
			return false
		}
		return vi.LessThan(vj)
	})

	resourceNameToVersions := make(map[string][]*model.Version)

	for i, v := range versions {
		for _, res := range v.Resources {
			// only generate files for the resources in our group, otherwise we import
			if !v.VersionConfig.IsOurProto(res.Filename) && !res.IsCustom {
				log.Printf("not generating solo-kit "+
					"clients for resource %v.%v, "+
					"resource proto package must match v proto package %v", res.ProtoPackage, res.Name, v.ProtoPackage)
				continue
			} else if res.IsCustom && res.CustomResource.Imported {
				log.Printf("not generating solo-kit "+
					"clients for resource %v.%v, "+
					"custom resources from a different v are not generated", res.GoPackage, res.Name, v.VersionConfig.GoPackage)
				continue
			}

			if _, found := resourceNameToVersions[res.Name]; !found {
				resourceNameToVersions[res.Name] = make([]*model.Version, 0, len(versions)-i)
			}
			resourceNameToVersions[res.Name] = append(resourceNameToVersions[res.Name], v)
		}
	}

	soloKitProject.Conversions = getConversionsFromResourceProjects(resourceNameToVersions)

	fs, err := generateFilesForConversionConfig(soloKitProject)
	if err != nil {
		return nil, err
	}
	files = append(files, fs...)

	for i := range files {
		files[i].Content = fileHeader + files[i].Content
	}

	return files, nil
}

func getConversionsFromResourceProjects(resNameToProjects map[string][]*model.Version) []*model.Conversion {
	conversions := make([]*model.Conversion, 0, len(resNameToProjects))
	for resName, projects := range resNameToProjects {
		if len(projects) < 2 {
			continue
		}
		conversion := &model.Conversion{
			Name:     resName,
			Projects: getConversionProjects(projects),
		}
		conversions = append(conversions, conversion)
	}

	// Sort conversions by name so reordering diffs aren't introduced to the conversion files
	sort.SliceStable(conversions, func(i, j int) bool { return conversions[i].Name < conversions[j].Name })

	return conversions
}

func generateFilesForConversionConfig(soloKitProject *model.ApiGroup) (code_generator.Files, error) {
	var v code_generator.Files
	for name, tmpl := range map[string]*template.Template{
		"resource_converter.sk.go":   templates.ConverterTemplate,
		"resource_converter_test.go": templates.ConverterTestTemplate,
	} {
		content, err := generateConversionFile(soloKitProject, tmpl)
		if err != nil {
			return nil, errors.Wrapf(err, "internal error: processing template '%v' for resource list %v failed", tmpl.ParseName, name)
		}
		v = append(v, code_generator.File{
			Filename: name,
			Content:  content,
		})
	}

	testSuite := &model.TestSuite{
		PackageName: soloKitProject.ConversionGoPackageShort,
	}
	for suffix, tmpl := range map[string]*template.Template{
		"_suite_test.go": templates.SimpleTestSuiteTemplate,
	} {
		name := testSuite.PackageName + suffix
		content, err := generateTestSuiteFile(testSuite, tmpl)
		if err != nil {
			return nil, errors.Wrapf(err, "internal error: processing template '%v' for resource list %v failed", tmpl.ParseName, name)
		}
		v = append(v, code_generator.File{
			Filename: name,
			Content:  content,
		})
	}

	return v, nil
}

func generateConversionFile(soloKitProject *model.ApiGroup, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, soloKitProject); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generateTestSuiteFile(suite *model.TestSuite, tmpl *template.Template) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, suite); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getConversionProjects(projects []*model.Version) []*model.ConversionProject {
	conversionProjects := make([]*model.ConversionProject, 0, len(projects))
	for index := range projects {
		conversionProjects = append(conversionProjects, getConversionProject(index, projects))
	}
	return conversionProjects
}

func getConversionProject(index int, projects []*model.Version) *model.ConversionProject {
	var nextVersion, previousVersion string
	if index < len(projects)-1 {
		nextVersion = projects[index+1].VersionConfig.Version
	}
	if index > 0 {
		previousVersion = projects[index-1].VersionConfig.Version
	}

	return &model.ConversionProject{
		Version:         projects[index].VersionConfig.Version,
		GoPackage:       projects[index].VersionConfig.GoPackage,
		NextVersion:     nextVersion,
		PreviousVersion: previousVersion,
	}
}
