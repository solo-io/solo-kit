package templates

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/templateutils"
	"text/template"
)

var ProjectTestSuiteTemplate = template.Must(template.New("project_template").Funcs(templateutils.Funcs).Parse(`package {{ .Version }}

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test{{ upper_camel .Name }}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{ upper_camel .Name }} Suite")
}





`))
