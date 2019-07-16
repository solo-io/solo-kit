package templates

import (
	"text/template"
)

var SimpleTestSuiteTemplate = template.Must(template.New("project_template").Funcs(Funcs).Parse(`package {{ .PackageName }}_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test{{ upper_camel .PackageName }}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{ upper_camel .PackageName }} Suite")
}

`))
