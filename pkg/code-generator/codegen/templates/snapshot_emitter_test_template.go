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
	"bytes"
	"context"
	"os"
	"fmt"
	"time"

	{{ .Imports }}
	"k8s.io/client-go/kubernetes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

	type metadataGetter interface {
		GetMetadata() *core.Metadata
	}

	var (
		ctx 				context.Context
		namespace1, namespace2         string
		namespace3, namespace4         string
		namespace5, namespace6         string
		name1, name2        = "angela"+helpers.RandString(3), "bob"+helpers.RandString(3)
		name3, name4       = "susan" + helpers.RandString(3), "jim" + helpers.RandString(3)
		name5 = "melisa" + helpers.RandString(3)
		labels1 = map[string]string{"env": "test"}
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


	deleteNamespaces := func(ctx context.Context, kube kubernetes.Interface, namespaces ...string) {
		err := kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespaces...)
		Expect(err).NotTo(HaveOccurred())
	}	

	// getNewNamespaces is used to generate new namespace names, so that we do not have to wait
	// when deleting namespaces in runNamespacedSelectorsWithWatchNamespaces. Since 
	// runNamespacedSelectorsWithWatchNamespaces uses watchNamespaces set to namespace1 and 
	// namespace2, this will work. Because the emitter willl only be watching namespaces that are
	// labeled.
	getNewNamespaces := func() {
		namespace3 = helpers.RandString(8)
		namespace4 = helpers.RandString(8)
		namespace5 = helpers.RandString(8)
		namespace6 = helpers.RandString(8)
	}

	// getNewNamespaces1and2 is used to generate new namespaces for namespace 1 and 2.
	// used for the same reason as getNewNamespaces() above
	getNewNamespaces1and2 := func() {
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
	}

	getMapOfNamespaceResources := func(getList func(string) ([]metadataGetter,error)) map[string][]string {
		namespaces := []string{namespace1, namespace2, namespace3, namespace4, namespace5, namespace6}
		namespaceResources := make(map[string][]string, len(namespaces))
		for _, ns := range namespaces {
			list, _ := getList(ns)
			for _, snap := range list {
				snapMeta := snap.GetMetadata()
				if _, hit := namespaceResources[snapMeta.Namespace]; hit {
					namespaceResources[snap.GetMetadata().Namespace] = make([]string, 1)
				}
				namespaceResources[snapMeta.Namespace] = append(namespaceResources[snapMeta.Namespace], snapMeta.Name)
			}
		}
		return namespaceResources
	}

	findNonMatchingResources := func(matchList, findList []metadataGetter) map[string][]string {
		nonMatching := make(map[string][]string)
		for _, snap := range matchList {
			snapMeta := snap.GetMetadata()
			matched := false
			for _,pre := range findList {
				preMeta := pre.GetMetadata()
				if preMeta.Namespace == snapMeta.Namespace && preMeta.Name == snapMeta.Name {
					matched = true
					break
				}
			}
			if ! matched {
				if _, hit := nonMatching[snapMeta.Namespace]; hit {
					nonMatching[snap.GetMetadata().Namespace] = make([]string, 1)
				}
				nonMatching[snapMeta.Namespace] = append(nonMatching[snapMeta.Namespace], snapMeta.Name)
			}
		}
		return nonMatching
	}

	findMatchingResources := func(matchList, findList []metadataGetter) map[string][]string {
		matching := make(map[string][]string)
		for _, snap := range matchList {
			snapMeta := snap.GetMetadata()
			matched := false
			for _,pre := range findList {
				preMeta := pre.GetMetadata()
				if preMeta.Namespace == snapMeta.Namespace && preMeta.Name == snapMeta.Name {
					matched = true
					break
				}
			}
			if matched {
				if _, hit := matching[snapMeta.Namespace]; hit {
					matching[snap.GetMetadata().Namespace] = make([]string, 1)
				}
				matching[snapMeta.Namespace] = append(matching[snapMeta.Namespace], snapMeta.Name)
			}
		}
		return matching
	}

	getMapOfResources := func(listOfResources []metadataGetter) map[string][]string {
		resources := make(map[string][]string)
		for _, snap := range listOfResources {
			snapMeta := snap.GetMetadata()
			if _, hit := resources[snapMeta.Namespace]; hit {
				resources[snap.GetMetadata().Namespace] = make([]string, 1)
			}
			resources[snapMeta.Namespace] = append(resources[snapMeta.Namespace], snapMeta.Name)
		}
		return resources
	}

{{- range .Resources }}
{{- if not .ClusterScoped }}
	convert{{ .PluralName }}ToMetadataGetter := func(rl {{ .ImportPrefix }}{{ .Name }}List) []metadataGetter {
		listConv := make([]metadataGetter, len(rl))
		for i, r := range rl {
			listConv[i] = r
		}
		return listConv
	}
{{- end }}
{{- end }}

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
		var previous *{{ .GoName }}Snapshot

{{- range .Resources }}

		/*
			{{ .Name }}
		*/
		assertSnapshot{{ .PluralName }} := func(expect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List, unexpect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List) {
			drain:
				for {
					select {
					case snap = <-snapshots:
						previous = snap
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
						Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
	{{- else }}
						var expectedResources map[string][]string
						var unexpectedResource map[string][]string

						if previous != nil {
							expectedResources = findNonMatchingResources(convert{{ .PluralName }}ToMetadataGetter(expect{{ .PluralName }}), convert{{ .PluralName }}ToMetadataGetter(previous.{{ upper_camel .PluralName }}))
							unexpectedResource = findMatchingResources(convert{{ .PluralName }}ToMetadataGetter(unexpect{{ .PluralName }}), convert{{ .PluralName }}ToMetadataGetter(previous.{{ upper_camel .PluralName }}))
						} else {
							expectedResources = getMapOfResources(convert{{ .PluralName }}ToMetadataGetter(expect{{ .PluralName }}))
							unexpectedResource = getMapOfResources(convert{{ .PluralName }}ToMetadataGetter(unexpect{{ .PluralName }}))
						}
						getList := func (ns string) ([]metadataGetter, error) {
							l, err := {{ lower_camel .Name }}Client.List(ns, clients.ListOpts{})
							return convert{{ .PluralName }}ToMetadataGetter(l), err
						}
						namespaceResources := getMapOfNamespaceResources(getList)
						Fail(fmt.Sprintf("expected final snapshot before 10 seconds. expected \nExpected:\n%#v\n\nUnexpected:\n%#v\n\nnamespaces:\n%#v", expectedResources, unexpectedResource, namespaceResources))
	{{- end }}
					}
				}
		}	

{{- if .ClusterScoped }}

		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a } 
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

		{{ lower_camel .Name }}3a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace1, name3, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}3a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil) 

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

 		{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name4), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}4a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil) 

		{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name5, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}5a )
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

		{{ lower_camel .Name }}4a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}4b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace4, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}4a)
		{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4b }
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)  

		{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace3, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}5b, err := {{ lower_camel .Name }}Client.Write(New{{ .Name }}WithLabels(namespace4, name2, labels1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}5a)
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}5b)
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched) 

		for _, r := range {{ lower_camel .Name }}NotWatched {
			err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
		}

{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}3a }
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a}...)
		assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Namespace, {{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.GetMetadata().Namespace, {{ lower_camel .Name }}1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b}...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b, {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a}
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3a.GetMetadata().Namespace, {{ lower_camel .Name }}3a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}3b.GetMetadata().Namespace, {{ lower_camel .Name }}3b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}3a, {{ lower_camel .Name }}3b}...)
		{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}4a, {{ lower_camel .Name }}5a}			
		assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}4a.GetMetadata().Namespace, {{ lower_camel .Name }}4a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Namespace, {{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a, {{ lower_camel .Name }}5b}...)
		assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- end }}

		// clean up environment
		deleteNamespaces(ctx, kube, namespace3, namespace4)
		getNewNamespaces()

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

		kubeutils.DeleteNamespacesInParallelBlocking(ctx, kube, namespace1, namespace2)

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
			{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name5), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())

			assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }, nil)
	{{- else }}
			{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name5), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name5), clients.WriteOpts{Ctx: ctx})
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
			var previous *{{ .GoName }}Snapshot

{{- range .Resources }}

			/*
				{{ .Name }}
			*/

			assertSnapshot{{ .PluralName }} := func(expect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List, unexpect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List) {
				drain:
					for {
						select {
						case snap = <-snapshots:
							previous = snap
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
							Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
{{- else }}

							var buffer bytes.Buffer
							if previous != nil {
								for _, sn := range previous.{{ upper_camel .PluralName }} {
									buffer.WriteString(fmt.Sprintf("namespace: %v name: %v    ", sn.GetMetadata().Namespace, sn.GetMetadata().Name))	
									buffer.WriteByte('\n')
								}
							} else {
								buffer.WriteString("****** NO PREVIOUS SNAP ********")
							}
							namespaces := []string{namespace1,namespace2,namespace3,namespace4,namespace5,namespace6}
							for i, ns := range namespaces {
								buffer.WriteString(fmt.Sprintf("*********** %d::%v ***********", i, ns))
								list, _ := {{ lower_camel .Name }}Client.List(ns, clients.ListOpts{})
								for _, sn := range list {
									buffer.WriteString(fmt.Sprintf("namespace: %v name: %v   ", sn.GetMetadata().Namespace, sn.GetMetadata().Name))	
									buffer.WriteByte('\n')
								}
							}
							buffer.WriteString("********** EXPECTED *********")
							for _,snap := range expect{{ .PluralName }} {
								buffer.WriteString(fmt.Sprintf("namespace: %v name: %v    ", snap.GetMetadata().Namespace, snap.GetMetadata().Name))	
							}
							buffer.WriteString("********* UNEXPECTED ***********")
							for _,snap := range unexpect{{ .PluralName }}{
								buffer.WriteString(fmt.Sprintf("namespace: %v name: %v    ", snap.GetMetadata().Namespace, snap.GetMetadata().Name))	
							}

							Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", buffer.String()))
{{- end }}
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

{{- end }}

			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name2), clients.WriteOpts{Ctx: ctx})
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

			createNamespaces(ctx, kube, namespace5)
			createNamespaceWithLabel(ctx, kube, namespace6, labels1)

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name5), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}5a)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- else }}

			{{ lower_camel .Name }}5a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}5b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace6, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}5a)
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}5b)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			{{ lower_camel .Name }}7a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace5, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}7b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace6, name4), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}7a)
			{{ lower_camel .Name }}Watched = append({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}7b)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			for _, r := range {{ lower_camel .Name }}NotWatched {
				err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
			}

{{- end }}

{{- if .ClusterScoped }}

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}5a }
			{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}5a }
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}2a)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}5a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}5a)
			assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- else }}

			for _, r := range {{ lower_camel .Name }}Watched {
				err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, r)
			}
			assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- end }}

			// clean up environment
			deleteNamespaces(ctx, kube, namespace3, namespace4, namespace5, namespace6)
			getNewNamespaces()

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
				{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name2), clients.WriteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b}
				assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

{{- if .ClusterScoped }}
				err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
{{- else }}
				err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.GetMetadata().Namespace, {{ lower_camel .Name }}1a.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.GetMetadata().Namespace, {{ lower_camel .Name }}1b.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
{{- end }}

				{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b}
				assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

				deleteNamespaces(ctx, kube, namespace1, namespace2)

				getNewNamespaces1and2()
				createNamespaces(ctx, kube, namespace1, namespace2)
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

			// create namespaces
			createNamespaceWithLabel(ctx, kube, namespace3, labels1)
			createNamespaceWithLabel(ctx, kube, namespace4, labels1)

{{- if .ClusterScoped }}

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{}
			{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, nil)

			deleteNamespaces(ctx, kube, namespace3)	

			for _, r := range {{ lower_camel .Name }}Watched {
				err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				{{ lower_camel .Name }}NotWatched  = append({{ lower_camel .Name }}NotWatched , r)
			}
			assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- else }}

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace4, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			{{ lower_camel .Name }}NotWatched := {{ .ImportPrefix }}{{ .Name }}List{}
			{{ lower_camel .Name }}Watched := {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b }
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			deleteNamespaces(ctx, kube, namespace3)

			{{ lower_camel .Name }}Watched = {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2b }
			{{ lower_camel .Name }}NotWatched = append({{ lower_camel .Name }}NotWatched, {{ lower_camel .Name }}2a)
			assertSnapshot{{ .PluralName }}({{ lower_camel .Name }}Watched, {{ lower_camel .Name }}NotWatched)

			for _, r := range {{ lower_camel .Name }}Watched {
				err = {{ lower_camel .Name }}Client.Delete(r.GetMetadata().Namespace, r.GetMetadata().Name, clients.DeleteOpts{Ctx: ctx})
				Expect(err).NotTo(HaveOccurred())
				{{ lower_camel .Name }}NotWatched  = append({{ lower_camel .Name }}NotWatched , r)
			}
			assertSnapshot{{ .PluralName }}(nil, {{ lower_camel .Name }}NotWatched)

{{- end }}

			deleteNamespaces(ctx, kube, namespace4)
			getNewNamespaces()

{{- end }}{{/* end of with */}}
{{- end }}{{/* end of range */}}
		})

{{- $num_of_clients_supported := 0 }}
{{- range .Resources }}
{{ if not .ClusterScoped }}
{{ if .HasStatus }}
		{{ $num_of_clients_supported = inc $num_of_clients_supported }}
{{- end }}
{{- end }}
{{- end }}

{{ if ge $num_of_clients_supported 1 }}
		It("should be able to return a resource from a deleted namespace, after the namespace is re-created", func () {
			ctx := context.Background()
			err := emitter.Register()
			Expect(err).NotTo(HaveOccurred())

			snapshots, errs, err := emitter.Snapshots([]string{""}, clients.WatchOpts{
				Ctx:                ctx,
				RefreshRate:        time.Second,
				ExpressionSelector: labelExpression1,
			})
			Expect(err).NotTo(HaveOccurred())

			var snap *TestingSnapshot
			var previous *TestingSnapshot

{{- range .Resources }}
{{ if not .ClusterScoped }}
{{ if .HasStatus }}

{{/* no need for anything else, this only works on clients that have kube resource factories, this will not work on clients that have memory resource factories.*/}}

			/*
			{{ .Name }}
			*/
			assertSnapshot{{ .PluralName }} := func(expect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List, unexpect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List) {
			drain:
				for {
					select {
					case snap = <-snapshots:
						previous = snap
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
						var expectedResources map[string][]string
						var unexpectedResource map[string][]string

						if previous != nil {
							expectedResources = findNonMatchingResources(convert{{ .PluralName }}ToMetadataGetter(expect{{ .PluralName }}), convert{{ .PluralName }}ToMetadataGetter(previous.{{ upper_camel .PluralName }}))
							unexpectedResource = findMatchingResources(convert{{ .PluralName }}ToMetadataGetter(unexpect{{ .PluralName }}), convert{{ .PluralName }}ToMetadataGetter(previous.{{ upper_camel .PluralName }}))
						} else {
							expectedResources = getMapOfResources(convert{{ .PluralName }}ToMetadataGetter(expect{{ .PluralName }}))
							unexpectedResource = getMapOfResources(convert{{ .PluralName }}ToMetadataGetter(unexpect{{ .PluralName }}))
						}
						getList := func (ns string) ([]metadataGetter, error) {
							l, err := {{ lower_camel .Name }}Client.List(ns, clients.ListOpts{})
							return convert{{ .PluralName }}ToMetadataGetter(l), err
						}
						namespaceResources := getMapOfNamespaceResources(getList)
						Fail(fmt.Sprintf("expected final snapshot before 10 seconds. expected \nExpected:\n%#v\n\nUnexpected:\n%#v\n\nnamespaces:\n%#v", expectedResources, unexpectedResource, namespaceResources))
					}
				}
			}	

			createNamespaceWithLabel(ctx, kube, namespace3, labels1)

			{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name1), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a}, nil)

			deleteNamespaces(ctx, kube, namespace3)
			Eventually(func () bool {
				_, err = kube.CoreV1().Namespaces().Get(ctx, namespace3, metav1.GetOptions{})
				return apierrors.IsNotFound(err)
			}, 10*time.Second, 1 * time.Second).Should(BeTrue())
			createNamespaceWithLabel(ctx, kube, namespace3, labels1)

			{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace3, name2), clients.WriteOpts{Ctx: ctx})
			Expect(err).NotTo(HaveOccurred())
			assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a}, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a})

			deleteNamespaces(ctx, kube, namespace3)
			Eventually(func () bool {
				_, err = kube.CoreV1().Namespaces().Get(ctx, namespace3, metav1.GetOptions{})
				return apierrors.IsNotFound(err)
			}, 10*time.Second, 1 * time.Second).Should(BeTrue())
{{- end }}
{{- end }}
{{- end }}
		})
{{- end }}{{/* if $num_of_clients_supported */}}
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
