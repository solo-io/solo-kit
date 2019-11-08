package factory

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination=./mocks/cluster_client_factory.go -source cluster_client_factory.go -package mocks

type ClusterClientFactory interface {
	GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error)
}
