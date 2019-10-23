package multicluster

import (
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

type ClusterClientHandler interface {
	HandleNewClusterClient(cluster string, client clients.ResourceClient)
	HandleRemovedClusterClient(cluster string, client clients.ResourceClient)
}

type ClusterClientCache interface {
	handler.ClusterHandler
	ClientForCluster(cluster string) (clients.ResourceClient, bool)
}

type clusterCache struct {
	clientGetter   ClientGetter
	clientHandlers []ClusterClientHandler
	clients        map[string]clients.ResourceClient
	clientAccess   sync.RWMutex
}

func NewClusterClientCache(clientGetter ClientGetter, handlers ...ClusterClientHandler) ClusterClientCache {
	return &clusterCache{
		clientGetter:   clientGetter,
		clientHandlers: handlers,
		clients:        make(map[string]clients.ResourceClient),
		clientAccess:   sync.RWMutex{},
	}
}

func (c *clusterCache) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := c.clientGetter.GetClient(cluster, restConfig)
	if err != nil {
		return
	}

	if err := client.Register(); err != nil {
		return
	}

	c.clientAccess.Lock()
	c.clients[cluster] = client
	c.clientAccess.Unlock()

	for _, h := range c.clientHandlers {
		h.HandleNewClusterClient(cluster, client)
	}
}

func (c *clusterCache) ClusterRemoved(cluster string, restConfig *rest.Config) {
	client, ok := c.ClientForCluster(cluster)
	if !ok {
		return
	}

	c.clientAccess.Lock()
	delete(c.clients, cluster)
	c.clientAccess.Unlock()

	for _, h := range c.clientHandlers {
		h.HandleRemovedClusterClient(cluster, client)
	}
}

func (c *clusterCache) ClientForCluster(cluster string) (clients.ResourceClient, bool) {
	c.clientAccess.RLock()
	client, ok := c.clients[cluster]
	c.clientAccess.RUnlock()
	return client, ok
}
