package options

type Options struct {
	Name       string
	ConfigFile string
	Config     Config
	Generate   Generate
	Init       Init
	Vaidate    Validate
}

type Generate struct {
	CompileProtos bool
}

type Config struct {
	Input         string
	Output        string
	Docs          string
	Cli           string
	Root          string
	ProjectName   string
	Env           []string
	ResourceRoots []string
}

type Init struct {
	Resources []string
}

type Validate struct {
	All bool
}
