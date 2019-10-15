package deployment

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type deploymentResourceClientGetter struct {
	ctx context.Context
}

var _ multicluster.ClientGetter = &deploymentResourceClientGetter{}

func NewDeploymentResourceClientGetter(ctx context.Context) *deploymentResourceClientGetter {
	return &deploymentResourceClientGetter{ctx: ctx}
}

func (g *deploymentResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	deploymentCache, err := cache.NewKubeDeploymentCache(g.ctx, kube)
	if err != nil {
		return nil, err
	}
	return newResourceClient(kube, deploymentCache), nil
}
