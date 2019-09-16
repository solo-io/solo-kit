package kubeutils

import (
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("owner ref conversion", func() {

	var (
		skRef *core.Metadata_OwnerReference
		ref   metav1.OwnerReference
	)

	BeforeEach(func() {
		skRef = &core.Metadata_OwnerReference{
			ApiVersion: "v1",
			Kind:       "Pod",
			Name:       "test",
			Uid:        "uuid",
		}

		ref = metav1.OwnerReference{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       "test",
			UID:        "uuid",
		}
	})
	Context("kube -> solo-kit", func() {
		It("can copy over all string fields properly", func() {
			skRefs := copyKubernetesOwnerReferences([]metav1.OwnerReference{ref})
			Expect(skRefs).To(HaveLen(1))
			Expect(skRefs[0]).To(BeEquivalentTo(skRef))
		})

		It("can properly copy over nil boolean fields", func() {
			falseValue := false
			trueValue := true
			ref.Controller = &falseValue
			ref.BlockOwnerDeletion = &trueValue
			skRef.BlockOwnerDeletion = &types.BoolValue{
				Value: true,
			}
			skRef.Controller = &types.BoolValue{
				Value: false,
			}
			skRefs := copyKubernetesOwnerReferences([]metav1.OwnerReference{ref})
			Expect(skRefs).To(HaveLen(1))
			Expect(skRefs[0]).To(BeEquivalentTo(skRef))
		})
	})

	Context("solo-kit -> kube", func() {
		It("can copy over all string fields properly", func() {
			kubeRefs := copySoloKitOwnerReferences([]*core.Metadata_OwnerReference{skRef})
			Expect(kubeRefs).To(HaveLen(1))
			Expect(kubeRefs[0]).To(BeEquivalentTo(ref))
		})

		It("can properly copy over nil boolean fields", func() {
			falseValue := false
			trueValue := true
			ref.Controller = &falseValue
			ref.BlockOwnerDeletion = &trueValue
			skRef.BlockOwnerDeletion = &types.BoolValue{
				Value: true,
			}
			skRef.Controller = &types.BoolValue{
				Value: false,
			}
			kubeRefs := copySoloKitOwnerReferences([]*core.Metadata_OwnerReference{skRef})
			Expect(kubeRefs).To(HaveLen(1))
			Expect(kubeRefs[0]).To(BeEquivalentTo(ref))
		})
	})
})
