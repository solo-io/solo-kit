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
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/namespace"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/k8s-utils/kubeutils"
	"github.com/solo-io/solo-kit/test/util"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/rest"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
		resourceNamespaceLister resources.ResourceNamespaceLister
		kubeCache cache.KubeCoreCache
	)
	const (
		TIME_BETWEEN_MESSAGES = 5
	)
{{- range .Resources }}
	New{{ .Name }}WithLabels := func(namespace, name string, labels map[string]string) (*{{ .ImportPrefix }}{{ .Name }}) {
		resource := {{ .ImportPrefix }}New{{ .Name }}(namespace, name)
		resource.GetMetadata().Labels = labels
		return resource
	}
{{- end }}

	createNamespaces := func(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
		err := kubeutils.CreateNamespacesInParallel(ctx, kube, namespaces...)
		Expect(err).NotTo(HaveOccurred())
	}

	createNamespaceWithLabel := func(ctx context.Context, kube kubernetes.Interface, namespace string, labels map[string]string) {
		_, err := kube.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: labels,
			},
		}, metav1.CreateOptions{})	
		Expect(err).ToNot(HaveOccurred())
	}

	deleteNonDefaultKubeNamespaces := func(ctx context.Context, kube kubernetes.Interface) {
		// clean up your local environment
		namespaces, err := kube.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		Expect(err).ToNot(HaveOccurred())
		defaultNamespaces := map[string]bool{"kube-node-lease": true, "kube-public": true, "kube-system": true, "local-path-storage": true, "default": true}
		var namespacesToDelete []string
		for _,ns := range namespaces.Items {
			if _, hit := defaultNamespaces[ns.Name]; ! hit{
				namespacesToDelete = append(namespacesToDelete, ns.Name)
			}
		}
		err = kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespacesToDelete...)	
		Expect(err).ToNot(HaveOccurred())
	}

	deleteNamespaces := func(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
		err := kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespaces...)
		Expect(err).NotTo(HaveOccurred())
	}	

	runNamespacedSelectorsWithWatchNamespaces := func() {
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
		{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a } 
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- end }}

{{- if .ClusterScoped }}

		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace1, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}2a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b }...)
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- end }}

{{- if .ClusterScoped }}

		{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace1, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}3a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace1, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}3b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace2, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b }...)
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- end }}

		createNamespaceWithLabel(ctx, kube, namespace3, labels1)
		createNamespaces(ctx, kube, namespace4)

{{- if .ClusterScoped }}

		{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}4a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}4b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace4, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}4a)
		{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4b }
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}


{{- if .ClusterScoped }}

		{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}5a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}5b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}5a)
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}5b)
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

{{- if .ClusterScoped }}

		{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}6a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name3, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}6b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace4, name3, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}6a)
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}6b)
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

		createNamespaceWithLabel(ctx, kube, namespace5, labels1)
		createNamespaces(ctx, kube, namespace6)

{{- if .ClusterScoped }}

		{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace5, name2, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}7a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace5, name1, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}7b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace6, name1, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}7a)
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}7b)
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

{{- if .ClusterScoped }}

		{{ lower_camel .Name }}8a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace5, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}8a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}8a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace6, name2, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}8b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace6, name3, labels2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}8a, {{ lower_camel .Name }}8b }...)
		assertNoMessageSent()

		for _, r := range {{ lower_camel .Name }}NotWatched {
			err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
		}
		assertNoMessageSent()

{{- end }}


{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}2a,  {{ lower_camel .Name }}3a }
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a, {{ lower_camel .Name }}8a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Namespace, {{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.GetMetadata().Namespace, {{ lower_camel .Name }}1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b}...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b, {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b, {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a}...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a, {{ lower_camel .Name }}8a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Namespace, {{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2b.GetMetadata().Namespace, {{ lower_camel .Name }}2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b}...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b, {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}6a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a }...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a, {{ lower_camel .Name }}8a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Namespace, {{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3b.GetMetadata().Namespace, {{ lower_camel .Name }}3b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b}...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a}			
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}7a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a }...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}8a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Namespace, {{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Namespace, {{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b}...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a}			
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}8a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}8a }...)
		assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}6a.GetMetadata().Namespace, {{ lower_camel .Name }}6a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}7a.GetMetadata().Namespace, {{ lower_camel .Name }}7a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a}...)
		assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- end }}
{{- end }}
	}

	BeforeEach(func() {
		err := os.Setenv(statusutils.PodNamespaceEnvName, "default")
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		namespace3 = helpers.RandString(8)
		namespace4 = helpers.RandString(8)
		namespace5 = helpers.RandString(8)
		namespace6 = helpers.RandString(8)

		kube = helpers.MustKubeClient()
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		resourceNamespaceLister = namespace.NewKubeClientCacheResourceNamespaceLister(kube, kubeCache)

		createNamespaces(ctx, kube, namespace1, namespace2)

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
		emitter = New{{ .GoName }}Emitter({{ $clients }}, resourceNamespaceLister)
	})
	AfterEach(func() {
		err := os.Unsetenv(statusutils.PodNamespaceEnvName)
		Expect(err).NotTo(HaveOccurred())

		deleteNonDefaultKubeNamespaces(ctx, kube)

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

		It("should be able to track all resources that are on labeled namespaces", func() {
			runNamespacedSelectorsWithWatchNamespaces()
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


{{- range .Resources }}

			/*
				{{ .Name }}
			*/

{{- if (not .ClusterScoped) }}
			assertNo{{ .PluralName }}Sent := func() {
				drain:
					for {
						select {
						case snap = <-snapshots:
							if len(snap.{{ upper_camel .PluralName }}) == 0 {
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
{{- end }}

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
			{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }
			assertNo{{ .PluralName }}Sent()

{{- end }}

			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}2a ) 
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b}
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)			

{{- end }}

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace1, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a } 
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace1, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}3b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace2, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b}...)
			assertNo{{ .PluralName }}Sent()

{{- end }}

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}4a ) 
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}4b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}4b }...)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

			createNamespaces(ctx, kube, namespace5, namespace6)

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}5a)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}5b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace6, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b }...)
			assertNoMessageSent()

{{- end }}

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace5, name2, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}6a) 
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}6a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace5, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}6b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace6, name3, labels1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}6b}...)
			assertNoMessageSent()

{{- end }}


{{- if .ClusterScoped }}

			{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name3), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}7a) 
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}7b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace6, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a, {{ lower_camel .Name }}7b }...)
			assertNoMessageSent()

			for _, r := range {{ lower_camel .Name }}NotWatched {
				err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
			}
			assertNoMessageSent()

{{- end }}

{{- if .ClusterScoped }}

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}3a, {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a   }
			{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a   }
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}2a)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- else }}

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Namespace, {{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2b.GetMetadata().Namespace, {{ lower_camel .Name }}2b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b}...)
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}4b}
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

{{- if .ClusterScoped }}

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a }
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}3a)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a }
			{{ lower_camel .Name }}NotWatched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a, {{ lower_camel .Name }}7a }
			{{ lower_camel .Name }}NotWatched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}6a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a }
			{{ lower_camel .Name }}NotWatched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}6a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}7a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}7a }
			assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- else }}

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Namespace, {{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4b.GetMetadata().Namespace, {{ lower_camel .Name }}4b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}4b}...)
			assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- end }}
{{- end }}
		})
	})

	Context("Tracking resources on namespaces that are deleted", func () {
		It("Should not contain resources from a deleted namespace", func () {
			ctx := context.Background()
			err := emitter.Register()
			Expect(err).NotTo(HaveOccurred())

			snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
				Ctx:         ctx,
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

				{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b}
				assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

				deleteNamespaces(ctx, kube, namespace1, namespace2)
				{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b}
				assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)	
{{- end }}
		})

		It("Should not contain resources from a deleted namespace, that is filtered", func () {
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


{{- $length := len .Resources }}
{{- $last_entry := minus $length 1 }}
{{- range $i, $r := .Resources }}
{{ with $r }}
			/*
				{{ .Name }}
			*/

			assertNo{{ .PluralName }}Sent := func() {
				drain:
					for {
						select {
						case snap = <-snapshots:
							if len(snap.{{ upper_camel .PluralName }}) == 0 {
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
			{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }
			assertNo{{ .PluralName }}Sent()

{{- end }}

			deleteNamespaces(ctx, kube, namespace1, namespace2)
			assertNo{{ .PluralName }}Sent()

			// create namespaces
			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a}
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

{{- end }}

			deleteNamespaces(ctx, kube, namespace3)	
			{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)	

			createNamespaceWithLabel(ctx, kube, namespace5, labels1)

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a}
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

			deleteNamespaces(ctx, kube, namespace4)	
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}3a)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			deleteNamespaces(ctx, kube, namespace4)	
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}2b)
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a}	
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			for _, r := range {{ lower_camel .Name }}Watched {
				err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
			}
			assertNo{{ .PluralName }}Sent()

{{- end }}

			deleteNamespaces(ctx, kube, namespace5)

{{/* after each resource we have to create the namespaces that were just deleted */}}
{{if ne $last_entry $i }}
			createNamespaces(ctx, kube, namespace1, namespace2)
{{- end }}
{{- end }}{{/* end of with */}}
{{- end }}{{/* end of range */}}
		})
	})

	Context("use different resource namespace listers", func() {
		BeforeEach(func () {
			resourceNamespaceLister = namespace.NewKubeClientResourceNamespaceLister(kube)
			emitter = New{{ .GoName }}Emitter({{ $clients }}, resourceNamespaceLister)
		})

		It("Should work with the Kube Client Namespace Lister", func () {
			runNamespacedSelectorsWithWatchNamespaces()
		})
	})

})

`))
