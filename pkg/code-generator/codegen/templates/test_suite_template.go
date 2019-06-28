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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/solo-kit/test/testutils"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test{{ upper_camel .ProjectConfig.Name }}(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "{{ upper_camel .ProjectConfig.Name }} Suite")
}


var (	
	lock      *clusterlock.TestClusterLocker
	cfg       *rest.Config

	_ = SynchronizedAfterSuite(func() {}, func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return
		}
		var err error
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		clientset, err := apiexts.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		
		{{- range $uniqueCrds}}
		err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete("{{lowercase .}}", &metav1.DeleteOptions{})
		testutils.ErrorNotOccuredOrNotFound(err)
		{{- end}}
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	_ = SynchronizedBeforeSuite(func() []byte {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			return nil
		}
		var err error
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		clientset, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		lock, err = clusterlock.NewTestClusterLocker(clientset, clusterlock.Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		return nil
	}, func([]byte) {})

)


`))
