package templates

import (
	"text/template"
)

var ProjectTestSuiteTemplate = template.Must(template.New("project_template").Funcs(Funcs).Parse(`package {{ .ProjectConfig.Version }}

{{- $uniqueCrds := new_str_slice }}
{{- range .Resources}}
{{- if  ne .ProtoPackage ""}}
{{- $uniqueCrds := (append_str_slice $uniqueCrds  (printf "%v.%v"  .PluralName .ProtoPackage))}}
{{- end }}
{{- end }}
{{- $uniqueCrds := (unique $uniqueCrds)}}

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
	"github.com/solo-io/solo-kit/test/testutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test{{ upper_camel .ProjectConfig.Name }}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{ upper_camel .ProjectConfig.Name }} Suite")
}


var (
	cfg *rest.Config

	_ = SynchronizedAfterSuite(func() {}, func() {
		var err error
		err = os.Unsetenv(statusutils.PodNamespaceEnvName)
		Expect(err).NotTo(HaveOccurred())

		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return
		}
		ctx := context.Background()
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		clientset, err := apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		
		{{- range $uniqueCrds}}
		err = clientset.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, "{{lowercase .}}", metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		{{- end}}
	})

	_ = SynchronizedBeforeSuite(func() []byte {
		var err error
		err = os.Setenv(statusutils.PodNamespaceEnvName, "default")
		Expect(err).NotTo(HaveOccurred())

		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return nil
		}
		return nil
	}, func([]byte) {})

)


`))
