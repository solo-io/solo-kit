package secretconverter

import (
	"context"

	apiv1 "github.com/solo-io/solo-kit/api/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kubesecret"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	v1 "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"

	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	kubeConfigKey = "kubeconfig"
)

func KubeCfgFromSecret(kubeSecret *kubev1.Secret) (*v1.KubeConfig, error) {
	rawKubeConfig, ok := kubeSecret.Data[kubeConfigKey]
	if !ok {
		// not a kubeconfig secret
		return nil, kubesecret.NotOurResource
	}
	baseConfig, err := clientcmd.Load(rawKubeConfig)
	if err != nil {
		return nil, err
	}
	meta := kubeutils.FromKubeMeta(kubeSecret.ObjectMeta)
	return &v1.KubeConfig{KubeConfig: apiv1.KubeConfig{Metadata: meta, KubeConfig: *baseConfig}}, nil
}

func KubeConfigToSecret(meta metav1.ObjectMeta, kubeconfig *clientcmdapi.Config) (*kubev1.Secret, error) {
	rawKubeConfig, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return nil, err
	}
	return &kubev1.Secret{ObjectMeta: meta, Data: map[string][]byte{kubeConfigKey: rawKubeConfig}}, nil
}

type KubeConfigSecretConverter struct{}

func (t *KubeConfigSecretConverter) FromKubeSecret(ctx context.Context, rc *kubesecret.ResourceClient, secret *kubev1.Secret) (resources.Resource, error) {
	return KubeCfgFromSecret(secret)
}

func (t *KubeConfigSecretConverter) ToKubeSecret(ctx context.Context, rc *kubesecret.ResourceClient, resource resources.Resource) (*kubev1.Secret, error) {
	kc, ok := resource.(*v1.KubeConfig)
	if !ok {
		return nil, errors.Errorf("can only convert a *v1.KubeConfig, received %T", resource)
	}
	return KubeConfigToSecret(kubeutils.ToKubeMeta(kc.Metadata), &kc.KubeConfig.KubeConfig)
}
