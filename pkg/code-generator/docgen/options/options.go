package options

import "fmt"

type DocsOutput string

const (
	Markdown     DocsOutput = "markdown"
	Restructured            = "restructured"
	Hugo                    = "hugo"
)

type HugoOptions struct {
	DataDir string
	ApiDir  string
}

type DocsOptions struct {
	Output      DocsOutput
	HugoOptions *HugoOptions
}

func ValidateGenDocs(genDocs *DocsOptions) error {
	if genDocs.HugoOptions != nil && genDocs.Output != Hugo {
		return fmt.Errorf("must only specify HugoOptions for Hugo docs generation, currently generating docs for %v", genDocs.Output)
	}
	return nil
}

const (
	HugoProtoDataFile = "ProtoMap.yaml"
)
