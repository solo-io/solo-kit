package schemagen

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rotisserie/eris"

	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/writer"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/utils/pointer"

	"github.com/ghodss/yaml"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	v1beta1 = "apiextensions.k8s.io/v1beta1"
)

var (
	ApiVersionMismatch = func(expected, actual string) error {
		return eris.Errorf("Expected ApiVersion [%s] but found [%s]", expected, actual)
	}
)

type CrdWriter struct {
	fileWriter writer.FileWriter
}

func NewCrdWriter(crdDirectory string) *CrdWriter {
	return &CrdWriter{
		fileWriter: &writer.DefaultFileWriter{
			Root: crdDirectory,
			HeaderFromFilename: func(s string) string {
				return fmt.Sprintf("# %s\n\n", writer.DefaultFileHeader)
			},
		},
	}
}

func (c *CrdWriter) ApplyValidationSchemaToCRD(crd apiextv1beta1.CustomResourceDefinition, validationSchema *apiextv1beta1.CustomResourceValidation) error {
	crd.Spec.Validation = validationSchema
	// Setting PreserveUnknownFields to false ensures that objects with unknown fields are rejected.
	// This is deprecated and will default to false in future versions.
	crd.Spec.PreserveUnknownFields = pointer.BoolPtr(false)

	crdBytes, err := yaml.Marshal(crd)
	if err != nil {
		return err
	}

	return c.fileWriter.WriteFile(code_generator.File{
		Filename: getFilenameForCRD(crd),
		Content:  string(crdBytes),
	})
}

func getFilenameForCRD(crd apiextv1beta1.CustomResourceDefinition) string {
	return fmt.Sprintf("%s_%s_%s.yaml", crd.Spec.Group, crd.Spec.Version, crd.Spec.Names.Kind)
}

func getCRDsFromDirectory(crdDirectory string) ([]apiextv1beta1.CustomResourceDefinition, error) {
	var crds []apiextv1beta1.CustomResourceDefinition

	err := filepath.Walk(crdDirectory, func(crdFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if !(strings.HasSuffix(crdFile, ".yaml") || strings.HasSuffix(crdFile, ".yml")) {
			return nil
		}

		crdFromFile, err := getCRDFromFile(crdFile)
		if err != nil {
			log.Fatalf("failed to get crd from file: %v", err)
			return err
		}
		crds = append(crds, crdFromFile)

		// Continue traversing the output directory
		return nil
	})
	return crds, err
}

func getCRDFromFile(pathToFile string) (apiextv1beta1.CustomResourceDefinition, error) {
	crd := apiextv1beta1.CustomResourceDefinition{}

	r, err := os.Open(pathToFile)
	if err != nil {
		return crd, err
	}
	defer func() {
		err := r.Close()
		if err != nil {
			log.Fatalf("failed to close file [%s]. %v", pathToFile, err)
		}
	}()

	f := bufio.NewReader(r)
	decoder := kubeyaml.NewYAMLReader(f)

	doc, err := decoder.Read()
	if err != nil {
		return crd, err
	}
	chunk := bytes.TrimSpace(doc)

	err = yaml.Unmarshal(chunk, &crd)
	if err == nil && v1beta1 != crd.APIVersion {
		return crd, ApiVersionMismatch(v1beta1, crd.APIVersion)
	}

	return crd, err
}
