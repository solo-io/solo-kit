package templates

import (
	"text/template"
)

var ResourceClientTestTemplate = template.Must(template.New("resource_client_test").Funcs(Funcs).Parse(`
// go:build solokit

package {{ .Project.ProjectConfig.Version }}

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
)

var _ = Describe("{{ .Name }}Client", func() {
	var ctx context.Context
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
		Context("resource client backed by "+test.Description(), func() {
			var (
				client {{ .Name }}Client
				err    error
				name1, name2, name3 = "foo"+helpers.RandString(3), "boo"+helpers.RandString(3), "goo"+helpers.RandString(3)
			)


{{- if .ClusterScoped }}
{{/* cluster-scoped resources get no namespace, must delete individual resources*/}}
			BeforeEach(func() {
				ctx = context.Background()
				factory := test.Setup(ctx, "")
				client, err = New{{ .Name }}Client(ctx, factory)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				client.Delete(name1, clients.DeleteOpts{})
				client.Delete(name2, clients.DeleteOpts{})
				client.Delete(name3, clients.DeleteOpts{})
			})

			It("CRUDs {{ .Name }}s "+test.Description(), func() {
				{{ .Name }}ClientTest(client, name1, name2, name3)
			})
{{- else }}
{{/* non-cluster-scoped resources get a namespace and then the ns is deleted*/}}
			BeforeEach(func() {
				namespace = helpers.RandString(6)
				ctx = context.Background()
				factory := test.Setup(ctx, namespace)
				client, err = New{{ .Name }}Client(ctx, factory)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				test.Teardown(ctx, namespace)
			})

			It("CRUDs {{ .Name }}s "+test.Description(), func() {
				{{ .Name }}ClientTest(namespace, client, name1, name2, name3)
			})
{{- end }}
		})
	}
})


{{- if .ClusterScoped }}
func {{ .Name }}ClientTest(client {{ .Name }}Client, name1, name2, name3 string) {
{{- else }}
func {{ .Name }}ClientTest(namespace string, client {{ .Name }}Client, name1, name2, name3 string) {
{{- end }}
	testOffset := 1

	err := client.Register()
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	name := name1

{{- if .ClusterScoped }}
	input := New{{ .Name }}("", name)
{{- else }}
	input := New{{ .Name }}(namespace, name)
{{- end }}

	r1, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	_, err = client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsExist(err)).To(BeTrue())

	ExpectWithOffset(testOffset, r1).To(BeAssignableToTypeOf(&{{ .Name }}{}))
	ExpectWithOffset(testOffset, r1.GetMetadata().Name).To(Equal(name))

{{- if (not .ClusterScoped) }}
	ExpectWithOffset(testOffset, r1.GetMetadata().Namespace).To(Equal(namespace))
{{- end }}
	ExpectWithOffset(testOffset, r1.GetMetadata().ResourceVersion).NotTo(Equal(input.GetMetadata().ResourceVersion))
	ExpectWithOffset(testOffset, r1.GetMetadata().Ref()).To(Equal(input.GetMetadata().Ref()))
	{{- range .Fields }}
		{{- if and (not (eq .Name "metadata")) (not .IsOneof) }}
	ExpectWithOffset(testOffset, r1.{{ upper_camel .Name }}).To(Equal(input.{{ upper_camel .Name }}))
		{{- end }}
	{{- end }}

	_, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())

	resources.UpdateMetadata(input, func(meta *core.Metadata) {
		meta.ResourceVersion = r1.GetMetadata().ResourceVersion
	})
	r1, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())


{{- if .ClusterScoped }}
	read, err := client.Read(name, clients.ReadOpts{})
{{- else }}
	read, err := client.Read(namespace, name, clients.ReadOpts{})
{{- end }}
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, read).To(Equal(r1))


{{- if (not .ClusterScoped) }}
	_, err = client.Read("doesntexist", name, clients.ReadOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsNotExist(err)).To(BeTrue())
{{- end }}

	name = name2
	input = &{{ .Name }}{}

	input.SetMetadata(&core.Metadata{
		Name:      name,
{{- if (not .ClusterScoped) }}
		Namespace: namespace,
{{- end }}
	})

	r2, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())


{{- if .ClusterScoped }}
	list, err := client.List(clients.ListOpts{})
{{- else }}
	list, err := client.List(namespace, clients.ListOpts{})
{{- end }}
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, list).To(ContainElement(r1))
	ExpectWithOffset(testOffset, list).To(ContainElement(r2))


{{- if .ClusterScoped }}
	err = client.Delete("adsfw", clients.DeleteOpts{})
{{- else }}
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{})
{{- end }}
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsNotExist(err)).To(BeTrue())


{{- if .ClusterScoped }}
	err = client.Delete("adsfw", clients.DeleteOpts{
{{- else }}
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{
{{- end }}
		IgnoreNotExist: true,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())


{{- if .ClusterScoped }}
	err = client.Delete(r2.GetMetadata().Name, clients.DeleteOpts{})
{{- else }}
	err = client.Delete(namespace, r2.GetMetadata().Name, clients.DeleteOpts{})
{{- end }}
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	Eventually(func() {{ .Name }}List {
{{- if .ClusterScoped }}
		list, err = client.List(clients.ListOpts{})
{{- else }}
		list, err = client.List(namespace, clients.ListOpts{})
{{- end }}
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second * 10).Should(ContainElement(r1))
	Eventually(func() {{ .Name }}List {
{{- if .ClusterScoped }}
		list, err = client.List(clients.ListOpts{})
{{- else }}
		list, err = client.List(namespace, clients.ListOpts{})
{{- end }}
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second * 10).ShouldNot(ContainElement(r2))

{{- if .ClusterScoped }}
	w, errs, err := client.Watch(clients.WatchOpts{
{{- else }}
	w, errs, err := client.Watch(namespace, clients.WatchOpts{
{{- end }}
		RefreshRate: time.Hour,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	var r3 resources.Resource
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		defer GinkgoRecover()

		resources.UpdateMetadata(r2, func(meta *core.Metadata) {
			meta.ResourceVersion = ""
		})
		r2, err = client.Write(r2, clients.WriteOpts{})
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

		name = name3
		input = &{{ .Name }}{}
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		input.SetMetadata(&core.Metadata{
			Name:      name,
{{- if (not .ClusterScoped) }}
			Namespace: namespace,
{{- end }}
		})

		r3, err = client.Write(input, clients.WriteOpts{})
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	}()
	<-wait

	select {
	case err := <-errs:
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	case list = <-w:
	case <-time.After(time.Millisecond * 5):
		Fail("expected a message in channel")
	}

	go func() {
		defer GinkgoRecover()
		for {
			select {
			case err := <-errs:
				ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
			case <-time.After(time.Second / 4):
				return
			}
		}
	}()

	Eventually(w, time.Second*5, time.Second/10).Should(Receive(And(ContainElement(r1), ContainElement(r3), ContainElement(r3))))
}
`))
