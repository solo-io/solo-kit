package crd

import (
	"strconv"

	"github.com/solo-io/solo-kit/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/testutils"
	"golang.org/x/sync/errgroup"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("crd unit tests", func() {

	var (
		baseCrd Crd
	)

	BeforeEach(func() {
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
			Expect(helpers.AddCrd(baseCrd)).NotTo(HaveOccurred())
			err := helpers.AddCrd(baseCrd)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(helpers.VersionExistsError(baseCrd.Version.Version)))
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
				Expect(helpers.AddCrd(v)).NotTo(HaveOccurred())
			}
			Expect(helpers.GetCrds()).To(HaveLen(1))
			Expect(helpers.GetCrds()[0].Versions).To(HaveLen(3))
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
				Expect(helpers.AddCrd(v)).NotTo(HaveOccurred())
			}
			Expect(helpers.GetCrds()).To(HaveLen(3))
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
					return helpers.AddCrd(v)
				})
			}
			Expect(eg.Wait()).NotTo(HaveOccurred())
			Expect(helpers.GetCrds()).To(HaveLen(2))
			for _, v := range helpers.GetCrds() {
				Expect(v.Versions).To(HaveLen(10))
			}
		})
		It("can retrieve available crds", func() {
			Expect(helpers.AddCrd(baseCrd)).NotTo(HaveOccurred())
			_, err := helpers.GetMultiVersionCrd(baseCrd.GroupKind())
			Expect(err).NotTo(HaveOccurred())
			_, err = helpers.GetCrd(baseCrd.GroupVersionKind())
			Expect(err).NotTo(HaveOccurred())
		})
		It("will fail if crd isn't available", func() {
			Expect(helpers.AddCrd(baseCrd)).NotTo(HaveOccurred())
			_, err := helpers.GetMultiVersionCrd(schema.GroupKind{})
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(helpers.NotFoundError(schema.GroupKind{}.String())))
			gvk := baseCrd.GroupVersionKind()
			gvk.Version = "hello"
			_, err = helpers.GetCrd(gvk)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(helpers.NotFoundError(gvk.String())))
		})
	})

	Context("CRD registration", func() {
		It("will error out if the corresponding gvk is not present", func() {
			Expect(helpers.AddCrd(baseCrd)).NotTo(HaveOccurred())
			gvk := baseCrd.GroupVersionKind()
			gvk.Version = "hello"
			mvCrd, err := helpers.GetMultiVersionCrd(baseCrd.GroupKind())
			Expect(err).NotTo(HaveOccurred())
			_, err = helpers.GetKubeCrd(mvCrd, gvk)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(helpers.InvalidGVKError(gvk)))
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
				Expect(helpers.AddCrd(v)).NotTo(HaveOccurred())
			}
			mvCrd, err := helpers.GetMultiVersionCrd(baseCrd.GroupKind())
			Expect(err).NotTo(HaveOccurred())
			crd, err := helpers.GetKubeCrd(mvCrd, crds[2].GroupVersionKind())
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
