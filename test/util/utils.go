package util

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned/fake"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ClientForClientsetAndResource(clientset *fake.Clientset, cache kube.SharedCache, crd crd.Crd, res resources.InputResource, namespaces []string) *kube.ResourceClient {
	return kube.NewResourceClient(
		crd,
		clientset,
		cache,
		res,
		namespaces,
		0)
}

func MockClientForNamespace(cache kube.SharedCache, namespaces []string) *kube.ResourceClient {
	return kube.NewResourceClient(
		v1.MockResourceCrd,
		fake.NewSimpleClientset(v1.MockResourceCrd),
		cache,
		&v1.MockResource{},
		namespaces,
		0)
}

func CreateMockResource(cs *fake.Clientset, namespace, name, dumbFieldValue string) error {
	_, err := cs.ResourcesV1().Resources(namespace).Create(
		v1.MockResourceCrd.KubeResource(&v1.MockResource{
			Metadata:      core.Metadata{Name: name},
			SomeDumbField: dumbFieldValue,
		}))
	return err
}

func DeleteMockResource(cs *fake.Clientset, namespace, name string) error {
	return cs.ResourcesV1().Resources(namespace).Delete(name, &metav1.DeleteOptions{})
}

func CreateV2Alpha1MockResource(cs *fake.Clientset, namespace, name, dumbFieldValue string) error {
	_, err := cs.ResourcesV1().Resources(namespace).Create(
		v2alpha1.MockResourceCrd.KubeResource(&v2alpha1.MockResource{
			Metadata: core.Metadata{Name: name},
			WeStuckItInAOneof: &v2alpha1.MockResource_SomeDumbField{
				SomeDumbField: dumbFieldValue,
			},
		}))
	return err
}
