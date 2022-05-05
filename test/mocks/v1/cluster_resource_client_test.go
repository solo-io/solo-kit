// Code generated by solo-kit. DO NOT EDIT.

//go:build solokit
// +build solokit

package v1

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
)

var _ = Describe("ClusterResourceClient", func() {
	var ctx context.Context
	for _, test := range []typed.ResourceClientTester{
		&typed.KubeRcTester{Crd: ClusterResourceCrd},
	} {
		Context("resource client backed by "+test.Description(), func() {
			var (
				client              ClusterResourceClient
				err                 error
				name1, name2, name3 = "foo" + helpers.RandString(3), "boo" + helpers.RandString(3), "goo" + helpers.RandString(3)
			)

			BeforeEach(func() {
				ctx = context.Background()
				factory := test.Setup(ctx, "")
				client, err = NewClusterResourceClient(ctx, factory)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				client.Delete(name1, clients.DeleteOpts{})
				client.Delete(name2, clients.DeleteOpts{})
				client.Delete(name3, clients.DeleteOpts{})
			})

			It("CRUDs ClusterResources "+test.Description(), func() {
				ClusterResourceClientTest(client, name1, name2, name3)
			})
		})
	}
})

func ClusterResourceClientTest(client ClusterResourceClient, name1, name2, name3 string) {
	testOffset := 1

	err := client.Register()
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	name := name1
	input := NewClusterResource("", name)

	r1, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	_, err = client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsExist(err)).To(BeTrue())

	ExpectWithOffset(testOffset, r1).To(BeAssignableToTypeOf(&ClusterResource{}))
	ExpectWithOffset(testOffset, r1.GetMetadata().Name).To(Equal(name))
	ExpectWithOffset(testOffset, r1.GetMetadata().ResourceVersion).NotTo(Equal(input.GetMetadata().ResourceVersion))
	ExpectWithOffset(testOffset, r1.GetMetadata().Ref()).To(Equal(input.GetMetadata().Ref()))

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
	read, err := client.Read(name, clients.ReadOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, read).To(Equal(r1))

	name = name2
	input = &ClusterResource{}

	input.SetMetadata(&core.Metadata{
		Name: name,
	})

	r2, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	list, err := client.List(clients.ListOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, list).To(ContainElement(r1))
	ExpectWithOffset(testOffset, list).To(ContainElement(r2))
	err = client.Delete("adsfw", clients.DeleteOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsNotExist(err)).To(BeTrue())
	err = client.Delete("adsfw", clients.DeleteOpts{
		IgnoreNotExist: true,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	err = client.Delete(r2.GetMetadata().Name, clients.DeleteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	Eventually(func() ClusterResourceList {
		list, err = client.List(clients.ListOpts{})
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).Should(ContainElement(r1))
	Eventually(func() ClusterResourceList {
		list, err = client.List(clients.ListOpts{})
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).ShouldNot(ContainElement(r2))
	w, errs, err := client.Watch(clients.WatchOpts{
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
		input = &ClusterResource{}
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		input.SetMetadata(&core.Metadata{
			Name: name,
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
