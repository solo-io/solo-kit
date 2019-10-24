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
type ClientForClusterHandler interface {
	HandleNewClusterClient(cluster string, client clients.ResourceClient)
	HandleRemovedClusterClient(cluster string, client clients.ResourceClient)
}

// Stores clients for clusters as they are discovered by a config watcher.
// Implementation in this package allows registration of ClusterClientHandlers.
type ClusterClientManager interface {
	// Returns a client for the given cluster if one exists.
	ClientForCluster(cluster string) (client clients.ResourceClient, found bool)
}

type clusterClientManager struct {
	ctx            context.Context
	clientGetter   ClientGetter
	clientHandlers []ClientForClusterHandler
	clients        map[string]clients.ResourceClient
	clientAccess   sync.RWMutex
}

var _ ClusterClientManager = &clusterClientManager{}
var _ handler.ClusterHandler = &clusterClientManager{}

func NewClusterClientManager(ctx context.Context, clientGetter ClientGetter, handlers ...ClientForClusterHandler) *clusterClientManager {
	return &clusterClientManager{
		clientGetter:   clientGetter,
		clientHandlers: handlers,
		clients:        make(map[string]clients.ResourceClient),
		clientAccess:   sync.RWMutex{},
	}
}

func (c *clusterClientManager) ClusterAdded(cluster string, restConfig *rest.Config) {
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

func (c *clusterClientManager) ClusterRemoved(cluster string, restConfig *rest.Config) {
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

func (c *clusterClientManager) ClientForCluster(cluster string) (client clients.ResourceClient, found bool) {
	c.clientAccess.RLock()
	client, found = c.clients[cluster]
	c.clientAccess.RUnlock()
	return client, found
}
