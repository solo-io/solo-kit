package options

type Options struct {
	Name       string
	ConfigFile string
	Config     Config
	Init       Init
	Vaidate    Validate
}

type Config struct {
	Input       string
	Output      string
	Docs        string
	Root        string
	ProjectName string
	Env         []string
	Imports     []string
}

type Init struct {
	Resources [] string
}

type Validate struct {
	All bool
}
