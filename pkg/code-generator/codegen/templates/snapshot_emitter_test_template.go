package templates

import (
	"text/template"
)

var ResourceGroupEmitterTestTemplate = template.Must(template.New("resource_group_emitter_test").Funcs(Funcs).Parse(`// +build solokit

package {{ .Project.ProjectConfig.Version }}

{{- /* we need to know if the tests require a crd client or a regular clientset */ -}}
{{- $client_declarations := new_str_slice }}
{{- $clients := new_str_slice }}
{{- $namespaces := new_str_slice }}
{{- $need_kube_config := false }}
{{- range $index, $val :=  .Resources}}
{{- $clients := (append_str_slice $clients (printf "%vClient"  (lower_camel .Name))) }}
{{- $client_declarations := (append_str_slice $client_declarations (printf "%vClient %v%vClient"  (lower_camel .Name) .ImportPrefix .Name)) }}
{{- $namespaces := (append_str_slice $namespaces (printf "namespace%v"  ($index))) }}
{{- if .HasStatus }}
{{- $need_kube_config = true }}
{{- end}}
{{- end}}
{{- $client_declarations := (join_str_slice $client_declarations ", ") }}
{{- $clients := (join_str_slice $clients ", ") }}
{{- $namespaces := (join_str_slice $namespaces ", ") }}

import (
	"context"
	"os"
	"time"

	{{ .Imports }}
	"github.com/solo-io/solo-kit/pkg/utils/stringutils"
	"golang.org/x/sync/errgroup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/go-utils/kubeutils"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var _ = Describe("{{ upper_camel .Project.ProjectConfig.Version }}Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
{{- range $index, $value := .Resources }}
		namespace{{ $index }}          string
{{- end }}
{{- range $index, $value := .Resources }}
		name{{ $index }} = helpers.RandString(8)
{{- end }}
		cfg                *rest.Config
		emitter            {{ .GoName }}Emitter
		kube 			   kubernetes.Interface
{{- range .Resources }}
		{{ lower_camel .Name }}Client {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
	)

	BeforeEach(func() {
{{- range $index, $value := .Resources }}
		namespace{{ $index }} = helpers.RandString(8)
{{- end }}
		var err error
{{- if $need_kube_config }}
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
{{- end}}

		kube = kubernetes.NewForConfigOrDie(cfg)
		err = setup.CreateNamespacesInParallel(kube, {{ $namespaces }})
		Expect(err).NotTo(HaveOccurred())

{{- range .Resources }}
		// {{ .Name }} Constructor

{{- if .HasStatus }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeResourceClientFactory{
			Crd: {{ .ImportPrefix }}{{ .Name }}Crd,
			Cfg: cfg,
		    SharedCache: kuberc.NewKubeCache(context.TODO()),
		}
{{- else }}
		{{ lower_camel .Name }}ClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
{{- end }}

		{{ lower_camel .Name }}Client, err = {{ .ImportPrefix }}New{{ .Name }}Client({{ lower_camel .Name }}ClientFactory)
		Expect(err).NotTo(HaveOccurred())
{{- end}}
		emitter = New{{ .GoName }}Emitter({{ $clients }})
	})
	AfterEach(func() {
		err := setup.DeleteNamespacesInParallelBlocking(kube, {{ $namespaces}} )
		Expect(err).NotTo(HaveOccurred())
	})

	var getAllNamespaces = func() []string {
		return []string{
		{{- range $index, $value := .Resources }}
			namespace{{ $index }},
		{{- end }}
		}
	}

	var getAllNames = func() []string {
		return []string{
		{{- range $index, $value := .Resources }}
			name{{ $index }},
		{{- end }}
		}
	}

	var {{ lower_camel .GoName }}EmitterTest = func(watchNamespaces *clients.NamespacesByResourceWatcher) {
		var (
			namespaces []string
			ok bool
		)
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())
	
		snapshots, errs, err := emitter.Snapshots(watchNamespaces, clients.WatchOpts{
			Ctx: ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())
	
		var snap *{{ .GoName }}Snapshot
		
		if watchNamespaces == nil {
			watchNamespaces = clients.NewNamespacesByResourceWatcher()
		}
	
{{- range .Resources }}

		/*
			{{ .Name }}
		*/

		namespaces, ok = watchNamespaces.Get({{ lower_camel .Name }}Client.BaseWatcher())
		if !ok || namespaces == nil {
			namespaces = []string{""}
		}
		
		assertSnapshot{{ .PluralName }} := func(expect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List, unexpect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expect{{ .PluralName }} {
{{- if .ClusterScoped }}
						if _, err := snap.{{ upper_camel .PluralName }}.Find(expected.GetMetadata().Ref().Strings()); err != nil {
{{- else }}
						if _, err := snap.{{ upper_camel .PluralName }}.List().Find(expected.GetMetadata().Ref().Strings()); err != nil {
{{- end }}
							continue drain
						}
					}
					for _, unexpected := range unexpect{{ .PluralName }} {
{{- if .ClusterScoped }}
						if _, err := snap.{{ upper_camel .PluralName }}.Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
{{- else }}
						if _, err := snap.{{ upper_camel .PluralName }}.List().Find(unexpected.GetMetadata().Ref().Strings()); err == nil {
{{- end }}
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
{{- if .ClusterScoped }}
					nsList, _ := {{ lower_camel .Name }}Client.List(clients.ListOpts{})
					combined := {{ .ImportPrefix }}{{ upper_camel .PluralName }}ByNamespace{
						"": nsList,
					}
{{- else }}
					combined := make({{ .ImportPrefix }}{{ upper_camel .PluralName }}ByNamespace)
					for _, namespace := range namespaces {
						nsList, err := {{ lower_camel .Name }}Client.List(namespace, clients.ListOpts{})
						Expect(err).NotTo(HaveOccurred())
						combined[namespace] = nsList
					}
{{- end }}
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}

		var expected{{ .Name }}s {{ .ImportPrefix }}{{ .Name }}List
		for _, name := range getAllNames() {
{{- if (not .ClusterScoped) }}
			for _, namespace := range getAllNamespaces() {
{{ else }}
				namespace := ""
{{- end }}
				{{ lower_camel .Name }}, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace, name), clients.WriteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				if stringutils.ContainsString(namespace, namespaces) || (len(namespaces) == 1 && namespaces[0] == "") {
					expected{{ .Name }}s = append(expected{{ .Name }}s, {{ lower_camel .Name }})
				}
{{- if (not .ClusterScoped) }}
			}
{{- end }}
		}
		assertSnapshot{{ .PluralName }}(expected{{ .Name }}s, nil)

		for _, expectedVal := range expected{{ .Name }}s {
{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete(expectedVal.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})

{{ else }}
			err = {{ lower_camel .Name }}Client.Delete(expectedVal.GetMetadata().Namespace, expectedVal.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})

{{- end }}
			Expect(err).NotTo(HaveOccurred())
		}
		assertSnapshot{{ .PluralName }}(nil, expected{{ .Name }}s)

{{- end}}
	}


	It("tracks snapshots on changes to any resource", func() {
		namespaces := clients.NewNamespacesByResourceWatcher()
		
{{- range .Resources }}
{{- if (not .ClusterScoped) }}
		namespaces.Set({{ lower_camel .Name }}Client.BaseWatcher(), getAllNamespaces())
{{- end }}
{{- end }}

		{{ lower_camel .GoName }}EmitterTest(namespaces)

	})

	It("tracks snapshots on changes to different resources in different namespaces", func() {
		namespaces := clients.NewNamespacesByResourceWatcher()
		
{{- range $index, $val := .Resources }}
{{- if (not .ClusterScoped) }}
		namespaces.Set({{ lower_camel .Name }}Client.BaseWatcher(), []string{namespace{{ $index }}})
{{- end }}
{{- end }}

		{{ lower_camel .GoName }}EmitterTest(namespaces)

	})

	It("tracks snapshots on changes to any resource using AllNamespace", func() {
		{{ lower_camel .GoName }}EmitterTest(nil)
	})

	
})

`))
