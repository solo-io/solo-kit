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

	"github.com/ghodss/yaml"
	code_generator "github.com/solo-io/solo-kit/pkg/code-generator"
	"github.com/solo-io/solo-kit/pkg/code-generator/writer"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	v1 = "apiextensions.k8s.io/v1"
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

func (c *CrdWriter) ApplyValidationSchemaToCRD(crd apiextv1.CustomResourceDefinition, validationSchema *apiextv1.CustomResourceValidation) error {
	crd.Spec.Versions[0].Schema = validationSchema
	crdBytes, err := yaml.Marshal(crd)
	if err != nil {
		return err
	}

	return c.fileWriter.WriteFile(code_generator.File{
		Filename: getFilenameForCRD(crd),
		Content:  string(crdBytes),
	})
}

func getFilenameForCRD(crd apiextv1.CustomResourceDefinition) string {
	return fmt.Sprintf("%s_%s_%s.yaml", crd.Spec.Group, crd.Spec.Versions[0].Name, crd.Spec.Names.Kind)
}

func GetCRDsFromDirectory(crdDirectory string) ([]apiextv1.CustomResourceDefinition, error) {
	var crds []apiextv1.CustomResourceDefinition

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

		crdFromFile, err := GetCRDFromFile(crdFile)
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

func GetCRDFromFile(pathToFile string) (apiextv1.CustomResourceDefinition, error) {
	crd := apiextv1.CustomResourceDefinition{}

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
	if err == nil && v1 != crd.APIVersion {
		return crd, ApiVersionMismatch(v1, crd.APIVersion)
	}

	return crd, err
}
