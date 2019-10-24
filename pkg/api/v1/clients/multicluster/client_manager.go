package multicluster

import (
	"context"
	"sync"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

//go:generate mockgen -destination=./mocks/client_manager.go -source client_manager.go -package mocks

// Allows objects to register callbacks with a ClusterClientManager.
type ClusterClientHandler interface {
	HandleNewClusterClient(cluster string, client clients.ResourceClient)
	HandleRemovedClusterClient(cluster string, client clients.ResourceClient)
}

// Stores clients for clusters as they are discovered by a config watcher.
// Implementation in this package allows registration of ClusterClientHandlers.
type ClusterClientManager interface {
	handler.ClusterHandler
	// Returns a client for the given cluster if one exists.
	ClientForCluster(cluster string) (client clients.ResourceClient, found bool)
}

type clusterManager struct {
	ctx            context.Context
	clientGetter   ClientGetter
	clientHandlers []ClusterClientHandler
	clients        map[string]clients.ResourceClient
	clientAccess   sync.RWMutex
}

func NewClusterClientManager(ctx context.Context, clientGetter ClientGetter, handlers ...ClusterClientHandler) ClusterClientManager {
	return &clusterManager{
		clientGetter:   clientGetter,
		clientHandlers: handlers,
		clients:        make(map[string]clients.ResourceClient),
		clientAccess:   sync.RWMutex{},
	}
}

func (c *clusterManager) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := c.clientGetter.GetClient(cluster, restConfig)
	if err != nil {
		contextutils.LoggerFrom(c.ctx).Error("failed to get client for cluster",
			zap.String("cluster", cluster),
			zap.Any("restConfig", restConfig))
		return
	}

	if err := client.Register(); err != nil {
		contextutils.LoggerFrom(c.ctx).Errorw("failed to register client for cluster",
			zap.String("cluster", cluster),
			zap.String("kind", client.Kind()))
		return
	}

	c.clientAccess.Lock()
	c.clients[cluster] = client
	c.clientAccess.Unlock()

	for _, h := range c.clientHandlers {
		h.HandleNewClusterClient(cluster, client)
	}
}

func (c *clusterManager) ClusterRemoved(cluster string, restConfig *rest.Config) {
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

func (c *clusterManager) ClientForCluster(cluster string) (client clients.ResourceClient, found bool) {
	c.clientAccess.RLock()
	client, found = c.clients[cluster]
	c.clientAccess.RUnlock()
	return client, found
}
