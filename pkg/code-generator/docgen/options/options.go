package options

type DocsOutput string

const (
	Markdown     DocsOutput = "markdown"
	Restructured            = "restructured"
	Hugo                    = "hugo"
)

type DocsOptions struct {
	Output DocsOutput
}
