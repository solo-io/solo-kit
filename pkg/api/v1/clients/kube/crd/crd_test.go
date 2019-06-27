package crd

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("crd unit tests", func() {

	var (
		baseCrd Crd
	)

	BeforeEach(func() {
		registry = &Registry{}
		baseCrd = Crd{
			CrdMeta: CrdMeta{
				KindName: "crdkind",
				Group:    "crdgroup",
			},
			Version: Version{
				Version: "crdversion",
			},
		}
	})
	Context("registry tests", func() {
		It("Adding the same crd twice results in an error", func() {
			Expect(registry.AddCrd(baseCrd)).NotTo(HaveOccurred())
			err := registry.AddCrd(baseCrd)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(VersionExistsError(baseCrd.Version.Version)))
		})
		It("adding multiple crds of the same gk with different versions will work successfully", func() {
			versionString := "crdversion"
			crdNumber := 3
			crds := make([]Crd, crdNumber)
			for i := 0; i < crdNumber; i++ {
				crds[i] = baseCrd
				crds[i].Version.Version = versionString + strconv.Itoa(i)
			}
			for _, v := range crds {
				v := v
				Expect(registry.AddCrd(v)).NotTo(HaveOccurred())
			}
			Expect(registry.crds).To(HaveLen(1))
			Expect(registry.crds[0].Versions).To(HaveLen(3))
		})
		It("adding multiple crds of different gk will result in different combined CRDs", func() {
			groupString := "crdgroup"
			crdNumber := 3
			crds := make([]Crd, crdNumber)
			for i := 0; i < crdNumber; i++ {
				crds[i] = baseCrd
				crds[i].Group = groupString + strconv.Itoa(i)
			}
			for _, v := range crds {
				v := v
				Expect(registry.AddCrd(v)).NotTo(HaveOccurred())
			}
			Expect(registry.crds).To(HaveLen(3))
		})

		It("can add many different crds simultaneously in go routines", func() {
			crdNumber := 20
			crds := make([]Crd, crdNumber)
			for i := 0; i < crdNumber; i++ {
				crds[i] = baseCrd
				if i%2 == 0 {
					crds[i].Group = "2"
				}
				crds[i].Version.Version = strconv.Itoa(i)
			}
			eg := errgroup.Group{}
			for _, v := range crds {
				v := v
				eg.Go(func() error {
					return registry.AddCrd(v)
				})
			}
			Expect(eg.Wait()).NotTo(HaveOccurred())
			Expect(registry.crds).To(HaveLen(2))
			for _, v := range registry.crds {
				Expect(v.Versions).To(HaveLen(10))
			}
		})
		It("can retrieve avilable crds", func() {
			Expect(registry.AddCrd(baseCrd)).NotTo(HaveOccurred())
			_, err := registry.GetCombinedCrd(baseCrd.GroupKind())
			Expect(err).NotTo(HaveOccurred())
			_, err = registry.GetCrd(baseCrd.GroupVersionKind())
			Expect(err).NotTo(HaveOccurred())
		})
		It("will fail if crd isn't available", func() {
			Expect(registry.AddCrd(baseCrd)).NotTo(HaveOccurred())
			_, err := registry.GetCombinedCrd(schema.GroupKind{})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(NotFoundError(schema.GroupKind{}.String())))
			gvk := baseCrd.GroupVersionKind()
			gvk.Version = "hello"
			_, err = registry.GetCrd(gvk)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(NotFoundError(gvk.String())))
		})

	})
})
