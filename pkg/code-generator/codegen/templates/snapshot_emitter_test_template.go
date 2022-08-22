package templates

import (
	"text/template"
)

var ResourceGroupEmitterTestTemplate = template.Must(template.New("resource_group_emitter_test").Funcs(Funcs).Parse(`// +build solokit

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
	"os"
	"time"

	{{ .Imports }}
	"k8s.io/client-go/kubernetes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
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
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		ctx 				context.Context
		namespace1, namespace2         string
		namespace3, namespace4         string
		namespace5, namespace6         string
		name1, name2        = "angela"+helpers.RandString(3), "bob"+helpers.RandString(3)
		name3, name4       = "susan" + helpers.RandString(3), "jim" + helpers.RandString(3)
		labels1 = map[string]string{"env": "test"}
		labels2 = map[string]string{"env": "testenv", "owner": "foo"}
		labelExpression1 =  "env in (test)"
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

{{- range .Resources }}
	{{ .ImportPrefix }}New{{ .Name }}WithLabels := func(namespace, name string, labels map[string]string) (*{{ .ImportPrefix }}{{ .Name }}) {
		resource := {{ .ImportPrefix }}New{{ .Name }}(namespace, name)
		resource.Metadata.Labels = labels
		return resource
	}
{{- end }}

	createNamespaces := func(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
		err := kubeutils.CreateNamespacesInParallel(ctx, kube, namespace5, namespace6)
		Expect(err).NotTo(HaveOccurred())
		for _,ns := range namespaces {
			if _,hit := createdNamespaces[ns]; !hit {
				createdNamespaces[ns] = true
			}
		}
	}

	BeforeEach(func() {
		err := os.Setenv(statusutils.PodNamespaceEnvName, "default")
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
		createdNamespaces = make(map[string]bool)
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		namespace3 = helpers.RandString(8)
		namespace4 = helpers.RandString(8)
		namespace5 = helpers.RandString(8)
		namespace6 = helpers.RandString(8)
		kube = helpers.MustKubeClient()
		createNamespaces(ctx, kube, namespace1, namespace2)
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

		namespacesToDelete := []string{}
		for namespace,_ := range createdNamespaces {
			namespacesToDelete = append(namespacesToDelete, namespace)
		}
		err = kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespacesToDelete...)
		Expect(err).NotTo(HaveOccurred())

{{- range .Resources }}
{{- if .ClusterScoped }}
		{{ lower_camel .Name }}Client.Delete(name1, clients.DeleteOpts{})
		{{ lower_camel .Name }}Client.Delete(name2, clients.DeleteOpts{})
{{- end }}
{{- end }}
	})

	Context("Tracking watched namespaces", func () {
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

		It("should be able to track resources that are labeled on other namespaces", func() {
			ctx := context.Background()
			err := emitter.Register()
			Expect(err).NotTo(HaveOccurred())

			// There is an error here in the code.
			snapshots, errs, err := emitter.Snapshots([]string{namespace1, namespace2}, clients.WatchOpts{
				Ctx:         ctx,
				RefreshRate: time.Second,
				ExpressionSelector: labelExpression1,
			})
			Expect(err).NotTo(HaveOccurred())

			var snap *{{ .GoName }}Snapshot

			assertNoMessageSent := func() {
				for {
					select {
					case snap = <-snapshots:
						Fail("expected that no snapshots would be recieved " + log.Sprintf("%v", snap))	
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 5):
						// this means that we have not recieved any mocks that we are not expecting
						return
					}
				}
			}
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
			notWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a } 
			assertNoMessageSent()
{{- else }}
			{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }
			assertSnapshot{{ .PluralName }}(watched, nil)
{{- end }}


{{- if .ClusterScoped }}
			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace1, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a } 
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b }...)
			assertSnapshotMocks(watched, nil)
{{- end }}

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace1, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a }...) 
			assertSnapshot{{ .PluralName }}(watched, notWatched)			
{{- else }}
			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace1, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}3b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace2, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b }...)
			assertSnapshotMocks(watched, nil)
{{- end }}

			createNamespaces(ctx, kube, namespace3, namespace4)

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a }...)
			assertNoMessageSent()
{{- else }}
			{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}4b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}4b }
			assertNoMessageSent()
{{- end }}

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a }...) 
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}5b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b }...)
			assertSnapshotMocks(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace3, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a }...) 
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace3, name3, labels2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}6b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace4, name3, labels2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}6b }...)
			assertNoMessageSent()
{{- end }}

			createNamespaces(ctx, kube, namespace5, namespace6)

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace5, name2, labels2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a }...)
			assertNoMessageSent()
{{- else }}
			{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace5, name1, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}7b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace6, name1, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b }...)
			assertSnapshotMocks(watched, notWatched)	
{{- end }}


{{- if .ClusterScoped }}
			{{ lower_camel .Name }}8a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace5, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}8a }...) 
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			{{ lower_camel .Name }}8a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace5, name2, labels2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}8b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace6, name2, labels2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}8a, {{ lower_camel .Name }}8b }...)
			assertNoMessageSent()
{{- end }}

			for _, r := range notWatched {
				err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
			}
			assertNoMessageSent()

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched ={{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a }
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a  }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Namespace, {{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.GetMetadata().Namespace, {{ lower_camel .Name }}1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b}...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b,  {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b, {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a }...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a  }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Namespace, {{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2b.GetMetadata().Namespace, {{ lower_camel .Name }}2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b}...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b, {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a }...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a  }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Namespace, {{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3b.GetMetadata().Namespace, {{ lower_camel .Name }}3b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b}...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b, {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}6a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a }...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a  }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Namespace, {{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5b.GetMetadata().Namespace, {{ lower_camel .Name }}5b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b}...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}7a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a }...)
			assertSnapshot{{ .PluralName }}(nil, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}7a.GetMetadata().Namespace, {{ lower_camel .Name }}7a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}7b.GetMetadata().Namespace, {{ lower_camel .Name }}7b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b}...)
			assertSnapshot{{ .PluralName }}(nil, notWatched)
{{- end }}
{{- end }}
		})
	})

	Context("Tracking empty watched namespaces", func () {
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

		It("should be able to track resources only made with the matching labels", func() {
			ctx := context.Background()
			err := emitter.Register()
			Expect(err).NotTo(HaveOccurred())

			snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
				Ctx:         ctx,
				RefreshRate: time.Second,
				ExpressionSelector: labelExpression1,
			})
			Expect(err).NotTo(HaveOccurred())

			var snap *{{ .GoName }}Snapshot

			assertNoMessageSent := func() {
				for {
					select {
					case snap = <-snapshots:
						Fail("expected that no snapshots wouldbe recieved " + log.Sprintf("%v", snap))	
					case err := <-errs:
						Expect(err).NotTo(HaveOccurred())
					case <-time.After(time.Second * 5):
						// this means that we have not recieved any mocks that we are not expecting
						return
					}
				}
			}

			assertNoMatchingMocks := func() {
				drain:
					for {
						select {
						case snap = <-snapshots:
							if len(snap.Mocks) == 0 {
								continue drain
							}
							Fail("expected that no snapshots containing resources would be recieved " + log.Sprintf("%v", snap))	
						case err := <-errs:
							Expect(err).NotTo(HaveOccurred())
						case <-time.After(time.Second * 5):
							// this means that we have not recieved any mocks that we are not expecting
							return
						}
					}
			}
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
			notWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }
			assertNoMatchingMocks()
{{- else }}
			{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }
			assertNoMatchingMocks()
{{- end }}

			createNamespaces(ctx, kube, namespace3, namespace4)

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a }...) 
			assertNoMatchingMocks()
{{- else }}
			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b }
			assertNoMatchingMocks()
{{- end }}

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace1, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a } 
			assertSnapshotMocks(watched, notWatched)
{{- else }}
			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace1, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}3b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace2, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b }
			assertSnapshotMocks(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a }...) 
			assertSnapshotMocks(watched, notWatched)
{{- else }}
			{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}4b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}4b }...)
			assertSnapshotMocks(watched, notWatched)
{{- end }}

			createNamespaces(ctx, kube, namespace5, namespace6)

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a }...) 
			assertNoMatchingMocks()
{{- else }}
			{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}5b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace6, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b }...)
			assertNoMessageSent()
{{- end }}

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace5, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a }...) 
			assertSnapshotMocks(watched, notWatched)
{{- else }}
			{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace5, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}6b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}WithLabels(namespace6, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = append(watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}6b }...)
			assertSnapshotMocks(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name3), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a }...) 
			assertNoMessageSent()
{{- else }}
			{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}7b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace6, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notWatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b }...)
			assertNoMessageSent()
{{- end }}

			for _, r := range notWatched {
				err = mockResourceClient.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
			}
			assertNoMessageSent()

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}6a }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Namespace, {{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3b.GetMetadata().Namespace, {{ lower_camel .Name }}3b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b}...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}4b, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}6b }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Namespace, {{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4b.GetMetadata().Namespace, {{ lower_camel .Name }}4b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}4b}...)
			watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}6b }
			assertSnapshot{{ .PluralName }}(watched, notWatched)
{{- end }}

{{- if .ClusterScoped }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}6a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			assertSnapshot{{ .PluralName }}(nil, notWatched)
{{- else }}
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}6a.GetMetadata().Namespace, {{ lower_camel .Name }}6a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}6b.GetMetadata().Namespace, {{ lower_camel .Name }}6b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			notwatched = append(notWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}6b}...)
			assertSnapshot{{ .PluralName }}(nil, notWatched)
{{- end }}
{{- end }}
		})
	})
})

`))
