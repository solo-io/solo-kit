package secretconverter

import (
	"context"
	"github.com/solo-io/go-utils/errors"

	apiv1 "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"

	kubev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const KubeCfgType kubev1.SecretType = "solo.io/kubeconfig"

func KubeCfgFromSecret(secret *kubev1.Secret) (*v1.KubeConfig, error) {
	if secret.Type != KubeCfgType {
		// not a kubeconfig secret
		return nil, kubesecret.NotOurResource
	}
	var keys []string
	for k := range secret.Data {
		keys = append(keys, k)
	}
	if len(keys) != 1 {
		return nil, errors.Errorf("kubeconfig secret data must contain exactly one value")
	}
	// cluster name is set from the key the user uses for their kubeconfig
	cluster := keys[0]
	baseConfig, err := clientcmd.Load(secret.Data[cluster])
	if err != nil {
		return nil, err
	}
	meta := kubeutils.FromKubeMeta(secret.ObjectMeta)
	return &v1.KubeConfig{KubeConfig: apiv1.KubeConfig{Metadata: meta, Config: *baseConfig, Cluster: cluster}}, nil
}

func KubeConfigToSecret(kc *v1.KubeConfig) (*kubev1.Secret, error) {
	rawKubeConfig, err := clientcmd.Write(kc.Config)
	if err != nil {
		return nil, err
	}
	return &kubev1.Secret{
		ObjectMeta: kubeutils.ToKubeMeta(kc.Metadata),
		Type:       KubeCfgType,
		Data:       map[string][]byte{kc.Cluster: rawKubeConfig},
	}, nil
}

type KubeConfigSecretConverter struct{}

func (t *KubeConfigSecretConverter) FromKubeSecret(ctx context.Context, rc *kubesecret.ResourceClient, secret *kubev1.Secret) (resources.Resource, error) {
	return KubeCfgFromSecret(secret)
}

func (t *KubeConfigSecretConverter) ToKubeSecret(ctx context.Context, rc *kubesecret.ResourceClient, resource resources.Resource) (*kubev1.Secret, error) {
	kc, ok := resource.(*v1.KubeConfig)
	if !ok {
		return nil, kubesecret.NotOurResource
	}
	return KubeConfigToSecret(kc)
}
