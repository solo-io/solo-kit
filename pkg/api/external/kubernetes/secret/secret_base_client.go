package secret

import (
	"context"

	"github.com/solo-io/solo-kit/api/external/kubernetes/secret"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	secretClient "github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	kubev1 "k8s.io/api/core/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

func NewSecretClient(kube kubernetes.Interface, cache cache.KubeCoreCache) skkube.SecretClient {
	resourceClient, _ := secretClient.NewResourceClientWithSecretConverter(
		kube,
		&skkube.Secret{},
		cache,
		&kubeConverter{},
	)
	return skkube.NewSecretClientWithBase(resourceClient)
}
func FromKubeSecret(cm *kubev1.Secret) *skkube.Secret {

	podCopy := cm.DeepCopy()
	kubeSecret := secret.Secret(*podCopy)
	resource := &skkube.Secret{
		Secret: kubeSecret,
	}

	return resource
}

func ToKubeSecret(resource resources.Resource) (*kubev1.Secret, error) {
	cmResource, ok := resource.(*skkube.Secret)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to config-map-only client", resources.Kind(resource))
	}

	cm := kubev1.Secret(cmResource.Secret)

	return &cm, nil
}

type kubeConverter struct{}

func (cc *kubeConverter) FromKubeSecret(ctx context.Context, rc *kubesecret.ResourceClient, secret *kubev1.Secret) (resources.Resource, error) {
	return FromKubeSecret(secret), nil
}

func (cc *kubeConverter) ToKubeSecret(ctx context.Context, rc *kubesecret.ResourceClient, resource resources.Resource) (*kubev1.Secret, error) {
	return ToKubeSecret(resource)
}
