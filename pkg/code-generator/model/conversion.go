package model

// SOLO-KIT Descriptors from which code can be generated

type Conversion struct {
	Name     string
	Projects []*ConversionProject
}

type ConversionProject struct {
	Version         string
	NextVersion     string
	PreviousVersion string
	GoPackage       string
}
