package options

type DocsOutput string

const (
	Markdown     DocsOutput = "markdown"
	Restructured            = "restructured"
	Hugo                    = "hugo"
)

type HugoOptions struct {
	// Directory where solo-kit should write the cross-project data file
	// If your docs are located in <repo_root>/docs/content/ (which is the case when your Hugo .contentDir="content")
	// then this value will be "docs/data", which will map to <repo_root>/docs/data/
	DataDir string
	// Directory, relative to Hugo's .contentDir, where the generated docs for the apis are placed
	// For example, if your hugo root is <repo_root>/docs, your .contentDir="content" and you want your api docs
	// to be served from "my.website.com/api/<generated>/<api>/<path>"
	// then you should set solo-kit.json's docs_dir="/docs/content/api" and set ApiDir="api"
	ApiDir string
}

type DocsOptions struct {
	Output      DocsOutput
	HugoOptions *HugoOptions
}

const (
	HugoProtoDataFile = "ProtoMap.yaml"

	// The relevant extensions when determining how to link to a resource
	HugoResourceExtension    = ".sk"
	DefaultResourceExtension = ".sk.md"
)
