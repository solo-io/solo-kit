package options

type Options struct {
	Name       string
	ConfigFile string
	Config     Config
}

type Config struct {
	Input          string
	Output         string
	Docs           string
	Root           string
	ProjectName    string
	GogoImports    []string
	SoloKitImports []string
}
