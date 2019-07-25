package generic

import (
	"time"

	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
)

// Call within "It"
func TestCrudClient(namespace1, namespace2 string, client ResourceClient, opts clients.WatchOpts, callbacks ...Callback) {
	selectors := opts.Selector
	foo := "foo"
	input := v1.NewMockResource(namespace1, foo)
	data := "hello: goodbye"
	input.Data = data
	labels := map[string]string{"pickme": helpers.RandString(8)}
	// add individual selectors
	for key, value := range selectors {
		labels[key] = value
	}
	input.Metadata.Labels = labels

	err := client.Register()
	Expect(err).NotTo(HaveOccurred())

	// list with no resources should return empty list, not err
	list, err := client.List("", clients.ListOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(list).To(BeEmpty())

	r1, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	postWrite(callbacks, r1)

	_, err = client.Write(input, clients.WriteOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsExist(err)).To(BeTrue())

	Expect(r1).To(BeAssignableToTypeOf(&v1.MockResource{}))
	Expect(r1.GetMetadata().Name).To(Equal(foo))
	Expect(r1.GetMetadata().Namespace).To(Equal(namespace1))
	Expect(r1.GetMetadata().ResourceVersion).NotTo(Equal(""))
	Expect(r1.(*v1.MockResource).Data).To(Equal(data))

	// if exists and resource ver was not updated, error
	_, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	Expect(err).To(HaveOccurred())

	resources.UpdateMetadata(input, func(meta *core.Metadata) {
		meta.ResourceVersion = r1.GetMetadata().ResourceVersion
	})
	data = "asdf: qwer"
	input.Data = data

	oldRv := r1.GetMetadata().ResourceVersion

	r1, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	Expect(err).NotTo(HaveOccurred())

	read, err := client.Read(namespace1, foo, clients.ReadOpts{})
	Expect(err).NotTo(HaveOccurred())
	postRead(callbacks, read)

	// it should update the resource version on the new write
	Expect(read.GetMetadata().ResourceVersion).NotTo(Equal(oldRv))
	Expect(read).To(Equal(r1))

	_, err = client.Read("doesntexist", foo, clients.ReadOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotExist(err)).To(BeTrue())

	boo := "boo"
	input = &v1.MockResource{
		Data: data,
		Metadata: core.Metadata{
			Name:      boo,
			Namespace: namespace2,
			Labels:    selectors,
		},
	}
	r2, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())

	// with labels
	list, err = client.List("", clients.ListOpts{
		Selector: labels,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(list).To(ContainElement(r1))
	postList(callbacks, list)
	Expect(list).NotTo(ContainElement(r2))

	// without
	list, err = client.List("", clients.ListOpts{
		Selector: selectors,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(list).To(And(ContainElement(r1), ContainElement(r2)))
	postList(callbacks, list)

	// test resource version locking works
	resources.UpdateMetadata(r2, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	_, err = client.Write(r2, clients.WriteOpts{OverwriteExisting: true})
	Expect(err).To(HaveOccurred())

	err = client.Delete(namespace1, "adsfw", clients.DeleteOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotExist(err)).To(BeTrue())

	err = client.Delete(namespace1, "adsfw", clients.DeleteOpts{
		IgnoreNotExist: true,
	})
	Expect(err).NotTo(HaveOccurred())

	err = client.Delete(r2.GetMetadata().Namespace, r2.GetMetadata().Name, clients.DeleteOpts{})
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() resources.ResourceList {
		list, err = client.List(namespace1, clients.ListOpts{
			Selector: selectors,
		})
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).Should(ContainElement(r1))
	Eventually(func() resources.ResourceList {
		list, err = client.List(namespace1, clients.ListOpts{
			Selector: selectors,
		})
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).ShouldNot(ContainElement(r2))

	// watch works on all namespaces
	w, errs, err := client.Watch("", opts)
	Expect(err).NotTo(HaveOccurred())

	var r3 resources.Resource
	wait := make(chan struct{})
	go func() {
		defer GinkgoRecover()
		defer func() {
			close(wait)
		}()
		resources.UpdateMetadata(r2, func(meta *core.Metadata) {
			meta.ResourceVersion = ""
		})
		r2, err = client.Write(r2, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		input = &v1.MockResource{
			Data: data,
			Metadata: core.Metadata{
				Name:      "goo",
				Namespace: namespace1,
				Labels:    selectors,
			},
		}
		r3, err = client.Write(input, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	}()
	select {
	case <-wait:
	case <-time.After(time.Second * 5):
		Fail("expected wait to be closed before 5s")
	}

	list = nil
	after := time.After(time.Millisecond * 250)
Loop:
	for {
		select {
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		case list = <-w:
		case <-after:
			if list == nil {
				Fail("expected a message in channel")
			}
			break Loop
		}
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

	postList(callbacks, list)

	Expect(list).To(ConsistOf(r1, r2, r3))
}

func postList(callbacks []Callback, list resources.ResourceList) {
	for _, el := range list {
		postRead(callbacks, el)
	}
}

func postRead(callbacks []Callback, res resources.Resource) {
	for _, cb := range callbacks {
		if cb.PostReadFunc != nil {
			cb.PostReadFunc(res)
		}
	}
}

func postWrite(callbacks []Callback, res resources.Resource) {
	for _, cb := range callbacks {
		if cb.PostWriteFunc != nil {
			cb.PostWriteFunc(res)
		}
	}
}

type Callback struct {
	PostReadFunc  func(res resources.Resource)
	PostWriteFunc func(res resources.Resource)
}
