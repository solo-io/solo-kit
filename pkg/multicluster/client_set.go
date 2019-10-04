package multicluster

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/rest"
)

// TODO this should be a shared struct
type ClusterClient struct {
	Client  clients.ResourceClient
	Cluster string
	// TODO maybe don't export these
	Ctx    context.Context
	Cancel context.CancelFunc
}

// TODO this should be a shared interface
type ClientSet interface {
	Clients() chan *ClusterClient
}

type ccWrapper struct {
	cc          *ClusterClient
	typedClient v2alpha1.MockResourceClient
}

type MockResourceClientSet interface {
	ClusterHandler
	ClientSet
	// TODO generate strongly typed client here
	ClientFor(cluster string) (v2alpha1.MockResourceClient, error)
}

type mockResourceClientSet struct {
	clients      map[string]*ccWrapper
	clientStream chan *ClusterClient
}

// TODO support some config, e.g. shared cache
func NewMockResourceClientSet() *mockResourceClientSet {
	return &mockResourceClientSet{
		clients:      make(map[string]*ccWrapper),
		clientStream: make(chan *ClusterClient),
	}
}

func (c *mockResourceClientSet) Clients() chan *ClusterClient {
	return c.clientStream
}

func (c *mockResourceClientSet) ClientFor(cluster string) (v2alpha1.MockResourceClient, error) {
	if cc, ok := c.clients[cluster]; ok {
		return cc.typedClient, nil
	}
	return nil, errors.Errorf("DNE")
}

func (c *mockResourceClientSet) ClusterAdded(cluster string, restConfig *rest.Config) {
	// TODO generate, support other types of clients
	krc := &factory.KubeResourceClientFactory{
		Cluster: cluster,
		Crd:     v2alpha1.MockResourceCrd,
		Cfg:     restConfig,
		// TODO Pass in through opts to constructor
		SharedCache:        nil,
		SkipCrdCreation:    false,
		NamespaceWhitelist: nil,
		ResyncPeriod:       0,
	}
	client, err := v2alpha1.NewMockResourceClient(krc)
	if err != nil {
		return
	}
	if err := client.Register(); err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	cc := &ClusterClient{
		Client:  client.BaseClient(),
		Cluster: cluster,
		Ctx:     ctx,
		Cancel:  cancel,
	}
	wrapper := &ccWrapper{
		cc:          cc,
		typedClient: client,
	}
	// TODO handle cases where clients for the cluster already exist ?
	c.clients[cluster] = wrapper
	c.clientStream <- cc
}

func (c *mockResourceClientSet) ClusterRemoved(cluster string, restConfig *rest.Config) {
	// cancel context associated with the ClusterClient, remove it from the set
	wrapper, ok := c.clients[cluster]
	if !ok {
		return
	}
	wrapper.cc.Cancel()
	delete(c.clients, cluster)
}
