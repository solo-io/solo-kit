package crd

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("crd unit tests", func() {

	var (
		baseCrd Crd
	)

	BeforeEach(func() {
		registry = &crdRegistry{}
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
			Expect(registry.addCrd(baseCrd)).NotTo(HaveOccurred())
			err := registry.addCrd(baseCrd)
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
				Expect(registry.addCrd(v)).NotTo(HaveOccurred())
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
				Expect(registry.addCrd(v)).NotTo(HaveOccurred())
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
					return registry.addCrd(v)
				})
			}
			Expect(eg.Wait()).NotTo(HaveOccurred())
			Expect(registry.crds).To(HaveLen(2))
			for _, v := range registry.crds {
				Expect(v.Versions).To(HaveLen(10))
			}
		})
		It("can retrieve available crds", func() {
			Expect(registry.addCrd(baseCrd)).NotTo(HaveOccurred())
			_, err := registry.getMultiVersionCrd(baseCrd.GroupKind())
			Expect(err).NotTo(HaveOccurred())
			_, err = registry.getCrd(baseCrd.GroupVersionKind())
			Expect(err).NotTo(HaveOccurred())
		})
		It("will fail if crd isn't available", func() {
			Expect(registry.addCrd(baseCrd)).NotTo(HaveOccurred())
			_, err := registry.getMultiVersionCrd(schema.GroupKind{})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(NotFoundError(schema.GroupKind{}.String())))
			gvk := baseCrd.GroupVersionKind()
			gvk.Version = "hello"
			_, err = registry.getCrd(gvk)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(NotFoundError(gvk.String())))
		})
	})

	Context("CRD registration", func() {
		It("will error out if the corresponding gvk is not present", func() {
			Expect(registry.addCrd(baseCrd)).NotTo(HaveOccurred())
			gvk := baseCrd.GroupVersionKind()
			gvk.Version = "hello"
			mvCrd, err := registry.getMultiVersionCrd(baseCrd.GroupKind())
			Expect(err).NotTo(HaveOccurred())
			_, err = registry.getKubeCrd(mvCrd, gvk)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(InvalidGVKError(gvk)))
		})
		It("can build the proper crd from multiple versions", func() {
			crdNumber := 3
			crds := make([]Crd, crdNumber)
			for i := 0; i < crdNumber; i++ {
				crds[i] = baseCrd
				crds[i].Version.Version = strconv.Itoa(i)
			}
			for _, v := range crds {
				v := v
				Expect(registry.addCrd(v)).NotTo(HaveOccurred())
			}
			mvCrd, err := registry.getMultiVersionCrd(baseCrd.GroupKind())
			Expect(err).NotTo(HaveOccurred())
			crd, err := registry.getKubeCrd(mvCrd, crds[2].GroupVersionKind())
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Scope).To(Equal(v1beta1.NamespaceScoped))
			Expect(crd.Spec.Group).To(Equal(mvCrd.Group))
			Expect(crd.GetName()).To(Equal(mvCrd.FullName()))
			Expect(crd.Spec.Versions).To(HaveLen(3))
			for _, v := range crd.Spec.Versions {
				if v.Name == "2" {
					Expect(v.Storage).To(BeTrue())
					Expect(v.Served).To(BeTrue())
				} else {
					Expect(v.Storage).To(BeFalse())
					Expect(v.Served).To(BeFalse())
				}
			}
		})
	})
})
