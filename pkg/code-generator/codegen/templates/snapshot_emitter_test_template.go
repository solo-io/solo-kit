package templates

import (
	"text/template"
)

var ResourceGroupEmitterTestTemplate = template.Must(template.New("resource_group_emitter_test").Funcs(Funcs).Parse(`
// go:build solokit

package {{ .Project.ProjectConfig.Version }}

{{- /* we need to know if the tests require a crd client or a regular clientset */ -}}
{{- $clients := new_str_slice }}
{{- $need_kube_config := false }}
{{- range .Resources}}
{{- $clients := (append_str_slice $clients (printf "%vClient"  (lower_camel .Name))) }}
{{- if .HasStatus }}
{{- $need_kube_config = true }}
{{- end}}
{{- end}}
{{- $clients := (join_str_slice $clients ", ") }}

import (
	"context"
	"fmt"
	"os"
	"time"

	{{ .Imports }}
	"k8s.io/client-go/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/solo-kit/test/util"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/rest"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("{{ upper_camel .Project.ProjectConfig.Version }}Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		fmt.Print("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ctx 				context.Context
		namespace1          string
		namespace2          string
		name1, name2        = "angela"+helpers.RandString(3), "bob"+helpers.RandString(3)
{{- if $need_kube_config }}
		cfg                *rest.Config
		clientset		   *apiext.Clientset
{{- end}}
		kube                      kubernetes.Interface
		emitter            {{ .GoName }}Emitter
{{- range .Resources }}
		{{ lower_camel .Name }}Client {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
	)

	BeforeEach(func() {
		err := os.Setenv(statusutils.PodNamespaceEnvName, "default")
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		err = kubeutils.CreateNamespacesInParallel(ctx, kube, namespace1, namespace2)
		Expect(err).NotTo(HaveOccurred())
{{- if $need_kube_config }}
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())

		clientset, err = apiext.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
{{- end}}

{{- range .Resources }}
		// {{ .Name }} Constructor

{{- if .HasStatus }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeResourceClientFactory{
			Crd: {{ .ImportPrefix }}{{ .Name }}Crd,
			Cfg: cfg,
		    SharedCache: kuberc.NewKubeCache(context.TODO()),
		}

		err = helpers.AddAndRegisterCrd(ctx, {{ .ImportPrefix }}{{ .Name }}Crd, clientset)
		Expect(err).NotTo(HaveOccurred())

{{- else }}
		{{ lower_camel .Name }}ClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
{{- end }}

		{{ lower_camel .Name }}Client, err = {{ .ImportPrefix }}New{{ .Name }}Client(ctx, {{ lower_camel .Name }}ClientFactory)
		Expect(err).NotTo(HaveOccurred())
{{- end}}
		emitter = New{{ .GoName }}Emitter({{ $clients }})
	})
	AfterEach(func() {
		err := os.Unsetenv(statusutils.PodNamespaceEnvName)
		Expect(err).NotTo(HaveOccurred())

		err = kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace1, namespace2)
		Expect(err).NotTo(HaveOccurred())

{{- range .Resources }}
{{- if .ClusterScoped }}
		{{ lower_camel .Name }}Client.Delete(name1, clients.DeleteOpts{})
		{{ lower_camel .Name }}Client.Delete(name2, clients.DeleteOpts{})
{{- end }}
{{- end }}
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
						if _, err := snap.{{ upper_camel .PluralName }}.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpect{{ .PluralName }} {
						if _, err := snap.{{ upper_camel .PluralName }}.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
{{- if .ClusterScoped }}
					combined, _ := {{ lower_camel .Name }}Client.List(clients.ListOpts{})
{{- else }}
					nsList1, _ := {{ lower_camel .Name }}Client.List(namespace1, clients.ListOpts{})
					nsList2, _ := {{ lower_camel .Name }}Client.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
{{- end }}
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}	

{{- if .ClusterScoped }}
		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }, nil)
{{- else }}
		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, nil)
{{- end }}

{{- if .ClusterScoped }}
		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}2a }, nil)
{{- else }}
		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b,  {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b  }, nil)
{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a })
{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Namespace, {{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2b.GetMetadata().Namespace, {{ lower_camel .Name }}2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })
{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}(nil, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}2a })
{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Namespace, {{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.GetMetadata().Namespace, {{ lower_camel .Name }}1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}(nil, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b, {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })
{{- end }}
{{- end}}
	})

	It("tracks snapshots on changes to any resource using AllNamespace", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
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
						if _, err := snap.{{ upper_camel .PluralName }}.Find(expected.GetMetadata().Ref().Strings()); err != nil {
							continue drain
						}
					}
					for _, unexpected := range unexpect{{ .PluralName }} {
						if _, err := snap.{{ upper_camel .PluralName }}.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
{{- if .ClusterScoped }}
					combined, _ := {{ lower_camel .Name }}Client.List(clients.ListOpts{})
{{- else }}
					nsList1, _ := {{ lower_camel .Name }}Client.List(namespace1, clients.ListOpts{})
					nsList2, _ := {{ lower_camel .Name }}Client.List(namespace2, clients.ListOpts{})
					combined := append(nsList1, nsList2...)
{{- end }}
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}	

{{- if .ClusterScoped }}
		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }, nil)
{{- else }}
		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, nil)
{{- end }}

{{- if .ClusterScoped }}
		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}2a }, nil)
{{- else }}
		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b,  {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b  }, nil)
{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a })
{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Namespace, {{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2b.GetMetadata().Namespace, {{ lower_camel .Name }}2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })
{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}(nil, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}2a })
{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Namespace, {{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.GetMetadata().Namespace, {{ lower_camel .Name }}1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}(nil, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b, {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })
{{- end }}
{{- end}}
	})
})

`))
