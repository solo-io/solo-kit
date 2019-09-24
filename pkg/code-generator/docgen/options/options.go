package options

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

const (
	HugoProtoDataFile = "ProtoMap.yaml"
)
