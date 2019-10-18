package multicluster

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination=./mocks/resource_client_getter.go -source resource_client_getter.go -package mocks

type ClientGetter interface {
	GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error)
}
