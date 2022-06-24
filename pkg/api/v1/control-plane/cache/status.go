// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package cache

import (
	"sync"
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/types"
)

// priority set for Envoy as listed here in Docs https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#resource-warming
var DefaultPrioritySet = map[int][]string{
	0: {types.ClusterTypeV1, types.ClusterTypeV2, types.ClusterTypeV3, types.ListenerTypeV1, types.ListenerTypeV2, types.ListenerTypeV3},
}

// NodeHash computes string identifiers for Envoy nodes.
type NodeHash interface {
	// ID function defines a unique string identifier for the remote Envoy node.
	ID(node *envoy_config_core_v3.Node) string
}

// StatusInfo tracks the server state for the remote Envoy node.
// Not all fields are used by all cache implementations.
type StatusInfo interface {
	// GetNode returns the node metadata.
	GetNode() *envoy_config_core_v3.Node

	// GetNumWatches returns the number of open watches.
	GetNumWatches() int

	// GetLastWatchRequestTime returns the timestamp of the last discovery watch request.
	GetLastWatchRequestTime() time.Time
}

type statusInfo struct {
	// node is the constant Envoy node metadata.
	node *envoy_config_core_v3.Node

	// watches are response watches of the original requests. They are structured by priority. See DefaultPrioritySet for more info.
	watches *PrioritySortedStruct

	// the timestamp of the last watch request
	lastWatchRequestTime time.Time

	// mutex to protect the status fields.
	// should not acquire mutex of the parent cache after acquiring this mutex.
	mu sync.RWMutex
}

// ResponseWatch is a watch record keeping both the request and an open channel for the response.
type ResponseWatch struct {
	// Request is the original request for the watch.
	Request Request

	// Response is the channel to push response to.
	Response chan Response
}

func (rw ResponseWatch) GetPriority() string {
	return rw.Request.TypeUrl
}

// NewStatusInfo initializes a status info data structure.
func NewStatusInfo(node *envoy_config_core_v3.Node, prioritySet map[int][]string) *statusInfo {
	if prioritySet == nil {
		prioritySet = DefaultPrioritySet
	}
	out := statusInfo{
		node:    node,
		watches: NewPrioritySortedStruct(prioritySet),
	}
	return &out
}

func (info *statusInfo) GetNode() *envoy_config_core_v3.Node {
	info.mu.RLock()
	defer info.mu.RUnlock()
	return info.node
}

func (info *statusInfo) GetNumWatches() int {
	info.mu.RLock()
	defer info.mu.RUnlock()
	return info.watches.Len()
}

func (info *statusInfo) GetLastWatchRequestTime() time.Time {
	info.mu.RLock()
	defer info.mu.RUnlock()
	return info.lastWatchRequestTime
}
