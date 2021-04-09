package v1beta1

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/rotisserie/eris"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	apiVersion      = "apiextensions.k8s.io/v1beta1"
	filePermissions = 0644
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

func WriteCRDSpecToFile(crd apiextv1beta1.CustomResourceDefinition, pathToFile string) error {
	// marshal to an empty field in the output
	crd.Status = apiextv1beta1.CustomResourceDefinitionStatus{}

	fileBytes, err := yaml.Marshal(crd)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(pathToFile, fileBytes, filePermissions)
}

func WriteCRDListToFile(crdList []apiextv1beta1.CustomResourceDefinition, pathToFile string) error {
	var manifests []string

	for _, crd := range crdList {
		// marshal to an empty field in the output
		crd.Status = apiextv1beta1.CustomResourceDefinitionStatus{}
		crdBytes, err := yaml.Marshal(crd)
		if err != nil {
			return err
		}
		manifests = append(manifests, string(crdBytes))
	}

	fileOutput := strings.Join(manifests, "---\n")

	return ioutil.WriteFile(pathToFile, []byte(fileOutput), filePermissions)
}
