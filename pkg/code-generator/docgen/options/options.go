package options

type DocsOutput string

const (
	Markdown     DocsOutput = "markdown"
	Restructured            = "restructured"
)

type DocsOptions struct {
	Output DocsOutput
}
