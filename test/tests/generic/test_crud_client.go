package generic

import (
	"time"

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
func TestCrudClient(namespace string, client ResourceClient, refreshRate time.Duration, callbacks ...Callback) {
	foo := "foo"
	input := v1.NewMockResource(namespace, foo)
	data := "hello: goodbye"
	input.Data = data
	labels := map[string]string{"pick": "me"}
	input.Metadata.Labels = labels

	err := client.Register()
	Expect(err).NotTo(HaveOccurred())

	r1, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	postWrite(callbacks, r1)

	_, err = client.Write(input, clients.WriteOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsExist(err)).To(BeTrue())

	Expect(r1).To(BeAssignableToTypeOf(&v1.MockResource{}))
	Expect(r1.GetMetadata().Name).To(Equal(foo))
	writtenNamespace := namespace
	if writtenNamespace == "" {
		writtenNamespace = DefaultNamespace
	}
	Expect(r1.GetMetadata().Namespace).To(Equal(writtenNamespace))
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

	read, err := client.Read(writtenNamespace, foo, clients.ReadOpts{})
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
			Namespace: namespace,
		},
	}
	r2, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())

	// with labels
	list, err := client.List(namespace, clients.ListOpts{
		Selector: labels,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(list).To(ContainElement(r1))
	postList(callbacks, list)
	Expect(list).NotTo(ContainElement(r2))

	// without
	list, err = client.List(namespace, clients.ListOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(list).To(ContainElement(r1))
	Expect(list).To(ContainElement(r2))
	postList(callbacks, list)

	err = client.Delete(writtenNamespace, "adsfw", clients.DeleteOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotExist(err)).To(BeTrue())

	err = client.Delete(writtenNamespace, "adsfw", clients.DeleteOpts{
		IgnoreNotExist: true,
	})
	Expect(err).NotTo(HaveOccurred())

	err = client.Delete(writtenNamespace, r2.GetMetadata().Name, clients.DeleteOpts{})
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() resources.ResourceList {
		list, err = client.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).Should(ContainElement(r1))
	Eventually(func() resources.ResourceList {
		list, err = client.List(namespace, clients.ListOpts{})
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).ShouldNot(ContainElement(r2))

	w, errs, err := client.Watch(namespace, clients.WatchOpts{RefreshRate: refreshRate})
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
				Namespace: namespace,
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

	select {
	case err := <-errs:
		Expect(err).NotTo(HaveOccurred())
	case list = <-w:
	case <-time.After(time.Millisecond * 5):
		Fail("expected a message in channel")
	}

	var timesDrained int
drain:
	for {
		select {
		case list = <-w:
			timesDrained++
			if timesDrained > 50 {
				Fail("drained the watch channel 50 times, something is wrong")
			}
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		case <-time.After(time.Second / 4):
			break drain
		}
	}

	postList(callbacks, list)
	Expect(list).To(ContainElement(r1))
	Expect(list).To(ContainElement(r2))
	Expect(list).To(ContainElement(r3))
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
