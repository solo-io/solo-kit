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
	"sort"
	"sync"
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

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

	// watches are indexed channels for the response watches and the original requests.
	watches map[int64]ResponseWatch

	// watchOrderingList is a list of watches optionally ordered by typeURL.
	// it is derived from watches via the orderResponseWatches function.
	watchOrderingList keys

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

// NewStatusInfo initializes a status info data structure.
func NewStatusInfo(node *envoy_config_core_v3.Node) *statusInfo {
	out := statusInfo{
		node:              node,
		watches:           make(map[int64]ResponseWatch),
		watchOrderingList: make(keys, 0),
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
	return len(info.watches)
}

func (info *statusInfo) GetLastWatchRequestTime() time.Time {
	info.mu.RLock()
	defer info.mu.RUnlock()
	return info.lastWatchRequestTime
}

// sortOrderingList to make sure that config can be sent in a non-conflicting
// order to the subscriber.
func (info *statusInfo) orderResponseWatches() {
	sort.Sort(info.watchOrderingList)
}

// generatewatchOrderingList places the watch map into a list with their
func (info *statusInfo) generatewatchOrderingList() {
	// 0 out our watch list cause watches get deleted in the map.
	info.watchOrderingList = make(keys, 0, len(info.watches))

	// This runs in O(n) which could become problematic when we have an extrclemely high watch count.
	// TODO(alec): revisit this and optimize for speed.
	for id, watch := range info.watches {
		info.watchOrderingList = append(info.watchOrderingList, key{
			ID:      id,
			TypeURL: watch.Request.TypeUrl,
		})
	}

}
