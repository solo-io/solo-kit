package options

type Options struct {
	Name       string
	ConfigFile string
	Config     Config
	Init
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
