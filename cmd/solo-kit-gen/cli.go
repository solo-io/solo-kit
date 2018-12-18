package solo_kit_gen

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cligen"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
)

func generateCli(projects []*model.Project, cliDir string) error {
	if len(projects) == 0 {
		return nil
	}

	cliProj := &model.CliProject{
		Resources:      []*model.Resource{},
		ResourceGroups: []*model.ResourceGroup{},
		CliConfig: model.CliConfig{
			Path: cliDir,
		},
	}

	// Combine all project Resources/Groups
	for _, proj := range projects {
		cliProj.Resources = append(cliProj.Resources, proj.Resources...)
		cliProj.ResourceGroups = append(cliProj.ResourceGroups, proj.ResourceGroups...)
	}

	code, err := cligen.GenerateFiles(cliProj, true)
	if err != nil {
		return err
	}

	err = writeFormatCode(code, cliDir)
	if err != nil {
		return err
	}

	return nil
}
