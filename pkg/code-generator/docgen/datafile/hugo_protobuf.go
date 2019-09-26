package datafile

// Hugo shortcodes can interpret and render data from

const HugoProtobufRelativeDataPath = "/data/ProtoMap.yaml"

type HugoProtobufData struct {
	Apis map[string]ApiSummary
}

type ApiSummary struct {
	// map from docs version to url relative to the Hugo .Site.BaseURL
	RelativePath string
	// protobuf package
	Package string
}
