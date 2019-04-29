package templates

import (
	"text/template"
)

var ResourceGroupEventLoopTestTemplate = template.Must(template.New("resource_group_event_loop_test").Funcs(Funcs).Parse(`// +build solokit

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
	"time"
	"sync"

	{{ .Imports }}
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("{{ .GoName }}EventLoop", func() {
	var (
		namespace string
		err       error
		emitter            {{ .GoName }}Emitter
{{- range .Resources }}
		{{ lower_camel .Name }}Client {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
	)

	BeforeEach(func() {
{{- range .Resources}}

		{{ lower_camel .Name }}ClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		{{ lower_camel .Name }}Client, err = {{ .ImportPrefix }}New{{ .Name }}Client({{ lower_camel .Name }}ClientFactory)
		Expect(err).NotTo(HaveOccurred())
{{- end}}

		emitter = New{{ .GoName }}Emitter({{ $clients }})
	})
	It("runs sync function on a new snapshot", func() {
{{- range .Resources  }}
		{{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
{{- end}}
		sync := &mock{{ .GoName }}Syncer{}
		el := New{{ .GoName }}EventLoop(emitter, sync)
		_, err := el.Run(nil, clients.WatchOpts{})
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
