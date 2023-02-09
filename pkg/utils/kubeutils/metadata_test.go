package kubeutils

import (
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("owner ref conversion", func() {

	var (
		skRef    *core.Metadata_OwnerReference
		ref      metav1.OwnerReference
		kubeMeta metav1.ObjectMeta
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

		kubeMeta = metav1.ObjectMeta{
			Name:            "test",
			Namespace:       "test",
			ResourceVersion: "1",
			Labels:          nil,
			Annotations:     nil,
			Generation:      1,
			OwnerReferences: []metav1.OwnerReference{ref},
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
			skRef.BlockOwnerDeletion = &wrappers.BoolValue{
				Value: true,
			}
			skRef.Controller = &wrappers.BoolValue{
				Value: false,
			}
			skRefs := copyKubernetesOwnerReferences([]metav1.OwnerReference{ref})
			Expect(skRefs).To(HaveLen(1))
			Expect(skRefs[0]).To(BeEquivalentTo(skRef))
		})

		It("can convert kube meta to core meta with owner references", func() {
			coreMeta := FromKubeMeta(kubeMeta, true)
			Expect(coreMeta).NotTo(BeNil())
			Expect(coreMeta.OwnerReferences).NotTo(BeNil())
			Expect(coreMeta.OwnerReferences).To(HaveLen(1))
			Expect(coreMeta.OwnerReferences[0]).To(BeEquivalentTo(skRef))
		})

		It("can convert kube meta to core meta without owner references", func() {
			coreMeta := FromKubeMeta(kubeMeta, false)
			Expect(coreMeta).NotTo(BeNil())
			Expect(coreMeta.OwnerReferences).To(BeNil())
			Expect(coreMeta.OwnerReferences).To(HaveLen(0))
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
			skRef.BlockOwnerDeletion = &wrappers.BoolValue{
				Value: true,
			}
			skRef.Controller = &wrappers.BoolValue{
				Value: false,
			}
			kubeRefs := copySoloKitOwnerReferences([]*core.Metadata_OwnerReference{skRef})
			Expect(kubeRefs).To(HaveLen(1))
			Expect(kubeRefs[0]).To(BeEquivalentTo(ref))
		})
	})
})
