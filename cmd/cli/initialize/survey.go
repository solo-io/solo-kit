package initialize

import (
	"fmt"
	"log"
	"strings"

	"github.com/solo-io/solo-kit/cmd/cli/options"
	"gopkg.in/AlecAivazis/survey.v1"
)

func basicInfoSurvey(cfg *options.Config) error {
	qsts := []*survey.Question{
		{
			Prompt: &survey.Input{
				Message: "Name for the project",
			},
			Name:     "projectname",
			Validate: survey.Required,
		},
		{
			Prompt: &survey.Input{
				Message: "Input directory",
			},
			Name:     "input",
			Validate: survey.Required,
		},
		{
			Prompt: &survey.Input{
				Message: "Output directory",
			},
			Name:     "output",
			Validate: survey.Required,
		},
		{
			Prompt: &survey.Input{
				Message: "Root directory (optional, defaults to dir containing solo-kit.yaml)",
			},
			Name: "root",
		},
	}
	return survey.Ask(qsts, cfg)
}

func resourceSurvey(cfg *options.Init) error {
	var resourcesString string
	qst := &survey.Input{
		Message: "Name(s) of resources. (comma seperated list: resource1,resource2,resource3)",
	}
	err := survey.AskOne(qst, &resourcesString, resourceSurveyCheck)
	if err != nil {
		return fmt.Errorf("resource name survey failed to complete")
	}
	resourceNames := strings.Split(resourcesString, ",")
	cfg.Resources = resourceNames
	return nil
}

func envSurvey(cfg *options.Config) error {
	var envString string
	qst := &survey.Input{
		Message: "Name(s) of env vars. (comma seperated list: var1=foo,var2=bar)",
	}
	err := survey.AskOne(qst, &envString, nil)
	if err != nil {
		log.Fatal("env vars survey failed to complete")
	}
	envVars := strings.Split(envString, ",")
	cfg.Env = envVars
	return nil
}

func resourceSurveyCheck(val interface{}) error {
	if err := survey.Required(val); err != nil {
		return err
	}
	resourceNames := strings.Split(val.(string), ",")
	if len(resourceNames) == 0 {
		return fmt.Errorf("resource list cannot be empty")
	}
	for _, v := range resourceNames {
		if len(v) < 3 {
			return fmt.Errorf("resource names must be at least 3 characters long")
		}
	}
	return nil
}
