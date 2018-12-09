package options

type Options struct {
	Name   string
	Config Config
}

type Config struct {
	Dir    string
	Input  string
	Output string
	Root   string
}
