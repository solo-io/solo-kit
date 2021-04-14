package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen"
	"github.com/solo-io/solo-kit/pkg/code-generator/schemagen/v1beta1"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

//go:generate go run generate.go
//go:generate ./api/v1/generate.sh
//go:generate ./pkg/api/v1/apiserver/generate.sh

func main() {

	log.Printf("starting generate")
	if err := cmd.Generate(cmd.GenerateOptions{
		RelativeRoot:       ".",
		CompileProtos:      true,
		SkipGenMocks:       true,
		SkipGeneratedTests: true,
		ExternalImports: &sk_anyvendor.Imports{
			Local: []string{"test/**/*.proto", "api/**/*.proto", sk_anyvendor.SoloKitMatchPattern},
			External: map[string][]string{
				sk_anyvendor.ExtProtoMatcher.Package:           sk_anyvendor.ExtProtoMatcher.Patterns,
				sk_anyvendor.EnvoyValidateProtoMatcher.Package: sk_anyvendor.EnvoyValidateProtoMatcher.Patterns,
			},
		},
		ValidationSchemaOpts: getValidationSchemaOptions(),
	}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}

func getValidationSchemaOptions() *schemagen.ValidationSchemaOptions {
	crd, _ := v1beta1.GetCRDFromFile("/pkg/code-generator/schemagen/v1beta1/fixtures/source/cc.yaml")

	return &schemagen.ValidationSchemaOptions{
		SchemaOptions: []*v1beta1.SchemaOptions{
			{
				OriginalCrd: crd,
				OnSchemaComplete: func(crdWithSchema apiextv1beta1.CustomResourceDefinition) error {
					// Do nothing
					return nil
				},
			},
		},
	}
}
