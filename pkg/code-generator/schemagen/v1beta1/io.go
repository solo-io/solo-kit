package v1beta1

import (
	"bufio"
	"bytes"
	"github.com/ghodss/yaml"
	"github.com/rotisserie/eris"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
	"os"
)

const (
	apiVersion = "apiextensions.k8s.io/v1beta1"
)

var (
	ApiVersionMismatch = func(expected, actual string) error {
		return eris.Errorf("Expected ApiVersion [%s] but found [%s]", expected, actual)
	}
)

func GetCRDFromFile(pathToFile string) (apiextv1beta1.CustomResourceDefinition, error) {
	crd := apiextv1beta1.CustomResourceDefinition{}

	r, err := os.Open(pathToFile)
	if err != nil {
		return crd, err
	}
	defer func() {
		_ = r.Close()
	}()

	f := bufio.NewReader(r)
	decoder := kubeyaml.NewYAMLReader(f)

	doc, err := decoder.Read()
	if err != nil {
		return crd, err
	}
	chunk := bytes.TrimSpace(doc)

	err = yaml.Unmarshal(chunk, &crd)
	if err == nil && apiVersion != crd.APIVersion {
		return crd, ApiVersionMismatch(apiVersion, crd.APIVersion)
	}

	return crd, err
}