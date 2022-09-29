package templates

import (
	"text/template"
)

var ResourceGroupEventLoopTestTemplate = template.Must(template.New("resource_group_event_loop_test").Funcs(Funcs).Parse(`// +build solokit

package {{ .Project.ProjectConfig.Version }}

{{- $clients := new_str_slice }}
{{- range .Resources}}
{{- $clients := (append_str_slice $clients (printf "%vClient" (lower_camel .Name))) }}
{{- end}}
{{- $clients := (join_str_slice $clients ", ") }}

import (
	"context"
	"time"
	"sync"

	{{ .Imports }}
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	skNamespace "github.com/solo-io/solo-kit/pkg/api/external/kubernetes/namespace"
	"github.com/solo-io/solo-kit/test/helpers"
)

var _ = Describe("{{ .GoName }}EventLoop", func() {
	var (
		ctx context.Context
		namespace string
		emitter     {{ .GoName }}Emitter
		err       error
	)

	BeforeEach(func() {
		ctx = context.Background()

		kube := helpers.MustKubeClient()
		kubeCache, err := cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		resourceNamespaceLister := skNamespace.NewKubeClientCacheResourceNamespaceLister(kube, kubeCache)

{{- range .Resources}}

		{{ lower_camel .Name }}ClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		{{ lower_camel .Name }}Client, err := {{ .ImportPrefix }}New{{ .Name }}Client(ctx, {{ lower_camel .Name }}ClientFactory)
		Expect(err).NotTo(HaveOccurred())
{{- end}}

		emitter = New{{ .GoName }}Emitter({{ $clients }}, resourceNamespaceLister)
	})
	It("runs sync function on a new snapshot", func() {
{{- range .Resources  }}
		_, err = emitter.{{ .Name }}().Write({{ .ImportPrefix }}New{{ .Name }}(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
{{- end}}
		sync := &mock{{ .GoName }}Syncer{}
		el := New{{ .GoName }}EventLoop(emitter, sync)
		_, err := el.Run([]string{namespace}, clients.WatchOpts{})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sync.Synced, 5*time.Second).Should(BeTrue())
	})
})

type mock{{ .GoName }}Syncer struct {
	synced bool
	mutex  sync.Mutex
}

func (s *mock{{ .GoName }}Syncer) Synced() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.synced
}

func (s *mock{{ .GoName }}Syncer) Sync(ctx context.Context, snap *{{ .GoName }}Snapshot) error {
	s.mutex.Lock()
	s.synced = true
	s.mutex.Unlock()
	return nil
}
`))
