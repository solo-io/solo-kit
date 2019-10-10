package templates

import (
	"text/template"
)

var MultiClusterResourceClientTestTemplate = template.Must(template.New("multi_cluster_client_test").Funcs(Funcs).Parse(`// +build solokit

package v1

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	v1 "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
)

var _ = Describe("{{ .Name }}MultiClusterClient", func() {
{{- if (not .ClusterScoped) }}
	var (
		namespace string
	)
{{- end }}
	for _, test := range []typed.ResourceClientTester{
{{- if (not .IsCustom) }}
		&typed.KubeRcTester{Crd: {{ .Name }}Crd},
{{- end }}
{{- /* cluster-scoped resources are currently only supported by crd client */}}
{{- if (not .ClusterScoped) }}
		&typed.ConsulRcTester{},
		&typed.FileRcTester{},
		&typed.MemoryRcTester{},
		&typed.VaultRcTester{},
		&typed.KubeSecretRcTester{},
		&typed.KubeConfigMapRcTester{},
{{- end }}
	} {
		Context("multi cluster resource client backed by "+test.Description(), func() {
			var (
				client              {{ .Name }}MultiClusterClient
				name1, name2, name3 = "foo" + helpers.RandString(3), "boo" + helpers.RandString(3), "goo" + helpers.RandString(3)
			)

{{- if .ClusterScoped }}
{{/* cluster-scoped resources get no namespace, must delete individual resources*/}}
			BeforeEach(func() {
				test.Setup("")
			})
			AfterEach(func() {
				client.Delete(name1, clients.DeleteOpts{})
				client.Delete(name2, clients.DeleteOpts{})
				client.Delete(name3, clients.DeleteOpts{})
			})
			It("CRUDs {{ .Name }}s "+test.Description(), func() {
				client = New{{ .Name }}MultiClusterClient(test)
				{{ .Name }}MultiClusterClientTest(client, name1, name2, name3)
			})
			It("errors when no client exists for the given cluster "+test.Description(), func() {
				client = New{{ .Name }}MultiClusterClient(test)
				{{ .Name }}MultiClusterClientCrudErrorsTest(client)
			})
			It("populates an aggregated watch "+test.Description(), func() {
				watchAggregator := wrapper.NewWatchAggregator()
				client = New{{ .Name }}MultiClusterClientWithWatchAggregator(watchAggregator, test)
				{{ .Name }}MultiClusterClientWatchAggregationTest(client, watchAggregator)
			})
{{- else }}
{{/* non-cluster-scoped resources get a namespace and then the ns is deleted*/}}
			BeforeEach(func() {
				namespace = helpers.RandString(6)
				test.Setup(namespace)
			})
			AfterEach(func() {
				test.Teardown(namespace)
			})
			It("CRUDs {{ .Name }}s "+test.Description(), func() {
				client = New{{ .Name }}MultiClusterClient(test)
				{{ .Name }}MultiClusterClientTest(namespace, client, name1, name2, name3)
			})
			It("errors when no client exists for the given cluster "+test.Description(), func() {
				client = New{{ .Name }}MultiClusterClient(test)
				{{ .Name }}MultiClusterClientCrudErrorsTest(client)
			})
			It("populates an aggregated watch "+test.Description(), func() {
				watchAggregator := wrapper.NewWatchAggregator()
				client = New{{ .Name }}MultiClusterClientWithWatchAggregator(watchAggregator, test)
				{{ .Name }}MultiClusterClientWatchAggregationTest(client, watchAggregator, namespace)
			})
{{- end }}
		})
	}
})

{{- if .ClusterScoped }}
func {{ .Name }}MultiClusterClientTest(client {{ .Name }}MultiClusterClient, name1, name2, name3 string) {
{{- else }}
func {{ .Name }}MultiClusterClientTest(namespace string, client {{ .Name }}MultiClusterClient, name1, name2, name3 string) {
{{- end }}
	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())
	client.ClusterAdded("", cfg)

	name := name1

{{- if .ClusterScoped }}
	input := New{{ .Name }}("", name)
{{- else }}
	input := New{{ .Name }}(namespace, name)
{{- end }}

	r1, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())

	_, err = client.Write(input, clients.WriteOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsExist(err)).To(BeTrue())

	Expect(r1).To(BeAssignableToTypeOf(&{{ .Name }}{}))
	Expect(r1.GetMetadata().Name).To(Equal(name))
{{- if (not .ClusterScoped) }}
	Expect(r1.GetMetadata().Namespace).To(Equal(namespace))
{{- end }}
	Expect(r1.GetMetadata().ResourceVersion).NotTo(Equal(input.GetMetadata().ResourceVersion))
	Expect(r1.GetMetadata().Ref()).To(Equal(input.GetMetadata().Ref()))
	{{- range .Fields }}
		{{- if and (not (eq .Name "metadata")) (not .IsOneof) }}
	Expect(r1.{{ upper_camel .Name }}).To(Equal(input.{{ upper_camel .Name }}))
		{{- end }}
	{{- end }}

	_, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	Expect(err).To(HaveOccurred())

	resources.UpdateMetadata(input, func(meta *core.Metadata) {
		meta.ResourceVersion = r1.GetMetadata().ResourceVersion
	})
	r1, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	Expect(err).NotTo(HaveOccurred())

{{- if .ClusterScoped }}
	read, err := client.Read(name, clients.ReadOpts{})
{{- else }}
	read, err := client.Read(namespace, name, clients.ReadOpts{})
{{- end }}
	Expect(err).NotTo(HaveOccurred())
	Expect(read).To(Equal(r1))

{{- if (not .ClusterScoped) }}
	_, err = client.Read("doesntexist", name, clients.ReadOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotExist(err)).To(BeTrue())
{{- end }}

	name = name2
	input = &{{ .Name }}{}

	input.SetMetadata(core.Metadata{
		Name:      name,
{{- if (not .ClusterScoped) }}
		Namespace: namespace,
{{- end }}
	})

	r2, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())

{{- if .ClusterScoped }}
	list, err := client.List(clients.ListOpts{})
{{- else }}
	list, err := client.List(namespace, clients.ListOpts{})
{{- end }}
	Expect(err).NotTo(HaveOccurred())
	Expect(list).To(ContainElement(r1))
	Expect(list).To(ContainElement(r2))
{{- if .ClusterScoped }}
	err = client.Delete("adsfw", clients.DeleteOpts{})
{{- else }}
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{})
{{- end }}
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotExist(err)).To(BeTrue())
{{- if .ClusterScoped }}
	err = client.Delete("adsfw", clients.DeleteOpts{
{{- else }}
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{
{{- end }}
		IgnoreNotExist: true,
	})
	Expect(err).NotTo(HaveOccurred())

{{- if .ClusterScoped }}
	err = client.Delete(r2.GetMetadata().Name, clients.DeleteOpts{})
{{- else }}
	err = client.Delete(namespace, r2.GetMetadata().Name, clients.DeleteOpts{})
{{- end }}
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() {{ .Name }}List {
{{- if .ClusterScoped }}
		list, err = client.List(clients.ListOpts{})
{{- else }}
		list, err = client.List(namespace, clients.ListOpts{})
{{- end }}
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).Should(ContainElement(r1))
	Eventually(func() {{ .Name }}List {
{{- if .ClusterScoped }}
		list, err = client.List(clients.ListOpts{})
{{- else }}
		list, err = client.List(namespace, clients.ListOpts{})
{{- end }}
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).ShouldNot(ContainElement(r2))
{{- if .ClusterScoped }}
	w, errs, err := client.Watch(clients.WatchOpts{
{{- else }}
	w, errs, err := client.Watch(namespace, clients.WatchOpts{
{{- end }}
		RefreshRate: time.Hour,
	})
	Expect(err).NotTo(HaveOccurred())

	var r3 resources.Resource
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		defer GinkgoRecover()

		resources.UpdateMetadata(r2, func(meta *core.Metadata) {
			meta.ResourceVersion = ""
		})
		r2, err = client.Write(r2, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		name = name3
		input = &{{ .Name }}{}
		Expect(err).NotTo(HaveOccurred())
		input.SetMetadata(core.Metadata{
			Name:      name,
{{- if (not .ClusterScoped) }}
			Namespace: namespace,
{{- end }}
		})

		r3, err = client.Write(input, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	}()
	<-wait

	select {
	case err := <-errs:
		Expect(err).NotTo(HaveOccurred())
	case list = <-w:
	case <-time.After(time.Millisecond * 5):
		Fail("expected a message in channel")
	}

	go func() {
		defer GinkgoRecover()
		for {
			select {
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case <-time.After(time.Second / 4):
				return
			}
		}
	}()

	Eventually(w, time.Second*5, time.Second/10).Should(Receive(And(ContainElement(r1), ContainElement(r3), ContainElement(r3))))
}

func {{ .Name }}MultiClusterClientCrudErrorsTest(client {{ .Name }}MultiClusterClient) {
{{- if .ClusterScoped }}
	_, err := client.Read("bar", clients.ReadOpts{Cluster: "read"})
{{- else }}
	_, err := client.Read("foo", "bar", clients.ReadOpts{Cluster: "read"})
{{- end }}
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(No{{ .Name }}ClientForClusterError("read").Error()))
{{- if .ClusterScoped }}
	_, err = client.List(clients.ListOpts{Cluster: "list"})
{{- else }}
	_, err = client.List("foo", clients.ListOpts{Cluster: "list"})
{{- end }}
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(No{{ .Name }}ClientForClusterError("list").Error()))
{{- if .ClusterScoped }}
	err = client.Delete("bar", clients.DeleteOpts{Cluster: "delete"})
{{- else }}
	err = client.Delete("foo", "bar", clients.DeleteOpts{Cluster: "delete"})
{{- end }}
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(No{{ .Name }}ClientForClusterError("delete").Error()))

	input = &{{ .Name }}{}
	input.SetMetadata(core.Metadata{
		Cluster:   "write",
		Name:      "bar",
{{- if (not .ClusterScoped) }}
		Namespace: namespace,
{{- end }}
	})
	_, err = client.Write(input, clients.WriteOpts{})
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(No{{ .Name }}ClientForClusterError("write").Error()))
{{- if .ClusterScoped }}
	_, _, err = client.Watch(clients.WatchOpts{Cluster: "watch"})
{{- else }}
	_, _, err = client.Watch("foo", clients.WatchOpts{Cluster: "watch"})
{{- end }}
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(Equal(No{{ .Name }}ClientForClusterError("watch").Error()))
}

func {{ .Name }}MultiClusterClientWatchAggregationTest(client {{ .Name }}MultiClusterClient, aggregator wrapper.WatchAggregator, namespace string) {
	w, errs, err := aggregator.Watch(namespace, clients.WatchOpts{})
	Expect(err).NotTo(HaveOccurred())
	go func() {
		defer GinkgoRecover()
		for {
			select {
			case err := <-errs:
				Expect(err).NotTo(HaveOccurred())
			case <-time.After(time.Second / 4):
				return
			}
		}
	}()

	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())
	client.ClusterAdded("", cfg)
	input = &{{ .Name }}{}
	input.SetMetadata(core.Metadata{
		Cluster:   "write",
		Name:      "bar",
{{- if (not .ClusterScoped) }}
		Namespace: namespace,
{{- end }}
	})
	_, err = client.Write(input, clients.WriteOpts{})
	written, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Eventually(w, time.Second*5, time.Second/10).Should(Receive(And(ContainElement(written))))
}
`))
