package utils

import "github.com/solo-io/solo-kit/pkg/code-generator/model"

func IsProjectResource(project *model.Project, resource *model.Resource) bool {
	return project.ProjectConfig.IsOurProto(resource.Filename)
}
