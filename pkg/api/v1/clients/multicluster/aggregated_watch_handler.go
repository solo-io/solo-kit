package multicluster

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
)

type aggregatedWatchClusterClientHandler struct {
	aggregator wrapper.WatchAggregator
}

var _ ClusterClientHandler = &aggregatedWatchClusterClientHandler{}

func NewAggregatedWatchClusterClientHandler(aggregator wrapper.WatchAggregator) ClusterClientHandler {
	return &aggregatedWatchClusterClientHandler{aggregator: aggregator}
}

func (h *aggregatedWatchClusterClientHandler) HandleNewClusterClient(cluster string, client clients.ResourceClient) {
	h.aggregator.AddWatch(client)
}

func (h *aggregatedWatchClusterClientHandler) HandleRemovedClusterClient(cluster string, client clients.ResourceClient) {
	h.aggregator.RemoveWatch(client)
}
