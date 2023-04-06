// Code generated by solo-kit. DO NOT EDIT.

//go:build solokit
// +build solokit

package kubernetes

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
)

var _ = Describe("ServiceClient", func() {
	var ctx context.Context
	var (
		namespace string
	)
	for _, test := range []typed.ResourceClientTester{
		&typed.ConsulRcTester{},
		&typed.FileRcTester{},
		&typed.MemoryRcTester{},
		&typed.VaultRcTester{},
		&typed.KubeSecretRcTester{},
		&typed.KubeConfigMapRcTester{},
	} {
		Context("resource client backed by "+test.Description(), func() {
			var (
				client              ServiceClient
				err                 error
				name1, name2, name3 = "foo" + helpers.RandString(3), "boo" + helpers.RandString(3), "goo" + helpers.RandString(3)
			)

			BeforeEach(func() {
				namespace = helpers.RandString(6)
				ctx = context.Background()
				factory := test.Setup(ctx, namespace)
				client, err = NewServiceClient(ctx, factory)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				test.Teardown(ctx, namespace)
			})

			It("CRUDs Services "+test.Description(), func() {
				ServiceClientTest(namespace, client, name1, name2, name3)
			})
		})
	}
})

func ServiceClientTest(namespace string, client ServiceClient, name1, name2, name3 string) {
	testOffset := 1

	err := client.Register()
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	name := name1
	input := NewService(namespace, name)

	r1, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	_, err = client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsExist(err)).To(BeTrue())

	ExpectWithOffset(testOffset, r1).To(BeAssignableToTypeOf(&Service{}))
	ExpectWithOffset(testOffset, r1.GetMetadata().Name).To(Equal(name))
	ExpectWithOffset(testOffset, r1.GetMetadata().Namespace).To(Equal(namespace))
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
	read, err := client.Read(namespace, name, clients.ReadOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, read).To(Equal(r1))
	_, err = client.Read("doesntexist", name, clients.ReadOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsNotExist(err)).To(BeTrue())

	name = name2
	input = &Service{}

	input.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})

	r2, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	list, err := client.List(namespace, clients.ListOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, list).To(ContainElement(r1))
	ExpectWithOffset(testOffset, list).To(ContainElement(r2))
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsNotExist(err)).To(BeTrue())
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{
		IgnoreNotExist: true,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	err = client.Delete(namespace, r2.GetMetadata().Name, clients.DeleteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	Eventually(func() ServiceList {
		list, err = client.List(namespace, clients.ListOpts{})
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).Should(ContainElement(r1))
	Eventually(func() ServiceList {
		list, err = client.List(namespace, clients.ListOpts{})
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).ShouldNot(ContainElement(r2))
	w, errs, err := client.Watch(namespace, clients.WatchOpts{
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
		input = &Service{}
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		input.SetMetadata(&core.Metadata{
			Name:      name,
			Namespace: namespace,
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
