package templates

import (
	"text/template"
)

var ResourceGroupEmitterTestTemplate = template.Must(template.New("resource_group_emitter_test").Funcs(funcs).Parse(`package {{ .Project.Version }}

{{- $clients := new_str_slice }}
{{- range .Resources}}
{{- $clients := (append_str_slice $clients (printf "%vClient"  (lower_camel .Name))) }}
{{- end}}
{{- $clients := (join_str_slice $clients ", ") }}

import (
	"context"
	"os"
	"path/filepath"
	"time"

	{{ .Imports }}
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("{{ upper_camel .Project.Version }}Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace1          string
		namespace2          string
		cfg                *rest.Config
		emitter            {{ .GoName }}Emitter
{{- range .Resources }}
		{{ lower_camel .Name }}Client {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
	)

	BeforeEach(func() {
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace1)
		Expect(err).NotTo(HaveOccurred())
		err = setup.SetupKubeForTest(namespace2)
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())

		cache := kuberc.NewKubeCache()

{{- range .Resources }}

		// {{ .Name }} Constructor
{{- if .HasStatus }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeResourceClientFactory{
			Crd: {{ .ImportPrefix }}{{ .Name }}Crd,
			Cfg: cfg,
		    SharedCache: cache,
		}
{{- else }}
{{/* TODO(ilackarms): Come with a more generic way to specify that a resource is "Secret"*/}}
		kube, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
{{- if (eq .Name "Secret") }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeSecretClientFactory{
			Clientset: kube,
		}
{{- else }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeConfigMapClientFactory{
			Clientset: kube,
		}
{{- end }}
{{- end }}
		{{ lower_camel .Name }}Client, err = {{ .ImportPrefix }}New{{ .Name }}Client({{ lower_camel .Name }}ClientFactory)
		Expect(err).NotTo(HaveOccurred())
{{- end}}
		emitter = New{{ .GoName }}Emitter({{ $clients }})
	})
	AfterEach(func() {
		setup.TeardownKube(namespace1)
		setup.TeardownKube(namespace2)
	})
	It("tracks snapshots on changes to any resource", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{namespace1, namespace2}, clients.WatchOpts{
			Ctx: ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *{{ .GoName }}Snapshot
{{- range .Resources }}

		/*
			{{ .Name }}
		*/
		
		assertSnapshot{{ .PluralName }} := func(expect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List, unexpect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expect{{ .PluralName }} {
						if _, err := snap.{{ .PluralName }}.List().Find(expected.Metadata.Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpect{{ .PluralName }} {
						if _, err := snap.{{ .PluralName }}.List().Find(unexpected.Metadata.Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
					nsList1, _ := {{ lower_camel .Name }}Client.List(namespace1, clients.ListOpts{})
					nsList2, _ := {{ lower_camel .Name }}Client.List(namespace2, clients.ListOpts{})
					combined := nsList1.ByNamespace()
					combined.Add(nsList2...)
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}	


		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, "angela"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, "angela"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, nil)

		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, "bob"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, "bob"), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b,  {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b  }, nil)

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.Metadata.Namespace, {{ lower_camel .Name }}2a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2b.Metadata.Namespace, {{ lower_camel .Name }}2b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.Metadata.Namespace, {{ lower_camel .Name }}1a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.Metadata.Namespace, {{ lower_camel .Name }}1b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}(nil, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b, {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })
{{- end}}
	})
})

`))
