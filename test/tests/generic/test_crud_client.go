package generic

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/matchers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
)

// TestCrudClient is called within "It" and performs a series of tests against a provided ResourceClient
func TestCrudClient(namespace1, namespace2 string, client ResourceClient, opts clients.WatchOpts, callbacks ...Callback) {
	testOffset := 1

	selectors := opts.Selector
	inputResourceName := "foo"
	input := v1.NewMockResource(namespace1, inputResourceName)
	data := "hello: goodbye"
	input.Data = data
	labels := map[string]string{"pickme": helpers.RandString(8)}
	// add individual selectors
	for key, value := range selectors {
		labels[key] = value
	}
	input.Metadata.Labels = labels

	err := client.Register()
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	// list with no resources should return empty list, not err
	list, err := client.List(namespace1, clients.ListOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, list).To(BeEmpty())

	// list with no resources should return empty list, not err
	list, err = client.List(namespace2, clients.ListOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, list).To(BeEmpty())

	r1, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	postWrite(callbacks, r1)

	_, err = client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsExist(err)).To(BeTrue())

	ExpectWithOffset(testOffset, r1).To(BeAssignableToTypeOf(&v1.MockResource{}))
	ExpectWithOffset(testOffset, r1.GetMetadata().Name).To(Equal(inputResourceName))
	ExpectWithOffset(testOffset, r1.GetMetadata().Namespace).To(Equal(namespace1))
	ExpectWithOffset(testOffset, r1.GetMetadata().ResourceVersion).NotTo(Equal(""))
	ExpectWithOffset(testOffset, r1.(*v1.MockResource).Data).To(Equal(data))

	// if exists and resource ver was not updated, error
	_, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())

	resources.UpdateMetadata(input, func(meta *core.Metadata) {
		meta.ResourceVersion = r1.GetMetadata().ResourceVersion
	})
	data = "asdf: qwer"
	input.Data = data

	oldRv := r1.GetMetadata().ResourceVersion

	r1, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	read, err := client.Read(namespace1, inputResourceName, clients.ReadOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	postRead(callbacks, read)

	// it should update the resource version on the new write
	ExpectWithOffset(testOffset, read.GetMetadata().ResourceVersion).NotTo(Equal(oldRv))
	ExpectWithOffset(testOffset, read).To(matchers.MatchProto(r1.(resources.ProtoResource)))

	_, err = client.Read("doesntexist", inputResourceName, clients.ReadOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsNotExist(err)).To(BeTrue())

	boo := "boo"
	input = &v1.MockResource{
		Data: data,
		Metadata: &core.Metadata{
			Name:      boo,
			Namespace: namespace2,
			Labels:    selectors,
		},
	}
	r2, err := client.Write(input, clients.WriteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	// with labels
	list, err = client.List("", clients.ListOpts{
		Selector: labels,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, list).To(matchers.ContainProto(r1.(resources.ProtoResource)))
	postList(callbacks, list)
	ExpectWithOffset(testOffset, list).NotTo(ContainElement(r2))

	// without
	list, err = client.List("", clients.ListOpts{
		Selector: selectors,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	ExpectWithOffset(testOffset, list).To(And(
		matchers.ContainProto(r1.(resources.ProtoResource)), matchers.ContainProto(r2.(resources.ProtoResource)),
	))
	postList(callbacks, list)

	// test resource version locking works
	resources.UpdateMetadata(r2, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	_, err = client.Write(r2, clients.WriteOpts{OverwriteExisting: true})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())

	err = client.Delete(namespace1, "adsfw", clients.DeleteOpts{})
	ExpectWithOffset(testOffset, err).To(HaveOccurred())
	ExpectWithOffset(testOffset, errors.IsNotExist(err)).To(BeTrue())

	err = client.Delete(namespace1, "adsfw", clients.DeleteOpts{
		IgnoreNotExist: true,
	})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	err = client.Delete(r2.GetMetadata().Namespace, r2.GetMetadata().Name, clients.DeleteOpts{})
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

	Eventually(func(g Gomega) resources.ResourceList {
		list, err = client.List(namespace1, clients.ListOpts{
			Selector: selectors,
		})
		g.ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).Should(matchers.ContainProto(r1.(resources.ProtoResource)))
	Eventually(func(g Gomega) resources.ResourceList {
		list, err = client.List(namespace1, clients.ListOpts{
			Selector: selectors,
		})
		g.ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
		return list
	}, time.Second*10).ShouldNot(ContainElement(r2))

	// watch works on all namespaces
	w, errs, err := client.Watch("", opts)
	ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

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
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())

		input = &v1.MockResource{
			Data: data,
			Metadata: &core.Metadata{
				Name:      "goo",
				Namespace: namespace1,
				Labels:    selectors,
			},
		}
		r3, err = client.Write(input, clients.WriteOpts{})
		ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
	}()
	select {
	case <-wait:
	case <-time.After(time.Second * 5):
		Fail("expected wait to be closed before 5s")
	}

	list = nil
	after := time.After(time.Second * 2)
Loop:
	for {
		select {
		case <-opts.Ctx.Done():
			return
		case err := <-errs:
			ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
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
			case <-opts.Ctx.Done():
				return
			case err := <-errs:
				ExpectWithOffset(testOffset, err).NotTo(HaveOccurred())
			case <-time.After(time.Second / 4):
				return
			}
		}
	}()

	postList(callbacks, list)

	ExpectWithOffset(testOffset, list).To(matchers.ConsistOfProtos(
		r1.(resources.ProtoResource),
		r2.(resources.ProtoResource),
		r3.(resources.ProtoResource)),
	)
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
