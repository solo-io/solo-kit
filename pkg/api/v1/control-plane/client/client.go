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

package client

import (
	"context"

	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/proto"

	envoy_api_v3_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"

	"google.golang.org/grpc"
)

type TypeRecord interface {
	Type() string
	EmptyProto() cache.ResourceProto
	ProtoToResource(r cache.ResourceProto) cache.Resource
}

type typeRecord struct {
	rtype           string
	proto           func() cache.ResourceProto
	protoToResource func(r cache.ResourceProto) cache.Resource
}

func NewTypeRecord(
	rtype string,
	proto func() cache.ResourceProto,
	protoToResource func(r cache.ResourceProto) cache.Resource,
) TypeRecord {
	return &typeRecord{
		rtype:           rtype,
		proto:           proto,
		protoToResource: protoToResource,
	}
}

func (t *typeRecord) Type() string {
	return t.rtype
}
func (t *typeRecord) EmptyProto() cache.ResourceProto {
	return t.proto()
}
func (t *typeRecord) ProtoToResource(r cache.ResourceProto) cache.Resource {
	return t.protoToResource(r)
}

type Client interface {
	Start(ctx context.Context, cc *grpc.ClientConn) error
}

type client struct {
	nodeinfo *envoy_api_v3_core.Node
	rtype    TypeRecord
	apply    func(cache.Resources) error
}

func NewClient(nodeinfo *envoy_api_v3_core.Node, rtype TypeRecord, apply func(cache.Resources) error) Client {
	return &client{
		nodeinfo: nodeinfo,
		rtype:    rtype,
		apply:    apply,
	}
}

/**
 * Start a client. this function is blocking.
 */
func (c *client) Start(ctx context.Context, cc *grpc.ClientConn) error {
	client := envoy_service_discovery_v3.NewAggregatedDiscoveryServiceClient(cc)
	resourceclient, err := client.StreamAggregatedResources(ctx)
	if err != nil {
		return err
	}
	// get a request going
	dr := &envoy_service_discovery_v3.DiscoveryRequest{
		VersionInfo:   "",
		Node:          c.nodeinfo,
		ResourceNames: []string{},
		TypeUrl:       c.rtype.Type(),
		ResponseNonce: "",
		ErrorDetail:   nil,
	}
	for {
		// make a copy of dr, to guarantee it doesnt get modified
		tosend := *dr
		err = resourceclient.Send(&tosend)
		if err != nil {
			return err
		}
		resp, err := resourceclient.Recv()
		if err != nil {
			return err
		}

		dr.ResponseNonce = resp.Nonce

		var resources cache.Resources
		resources.Version = resp.VersionInfo
		resources.Items = make(map[string]cache.Resource)
		for _, r := range resp.Resources {
			into := c.rtype.EmptyProto()
			err = proto.Unmarshal(r.Value, into)
			if err != nil {
				break
			}
			resource := c.rtype.ProtoToResource(into)
			resources.Items[resource.Self().Name] = resource
		}
		// If we have an error, don't update version info to signal NACK.
		if err != nil {
			dr.ErrorDetail = &status.Status{
				Code:    int32(codes.InvalidArgument),
				Message: err.Error(),
			}
		} else if err = c.apply(resources); err != nil {

			dr.ErrorDetail = &status.Status{
				Code:    int32(codes.Unknown),
				Message: err.Error(),
			}
		} else {
			dr.VersionInfo = resp.VersionInfo
			dr.ErrorDetail = nil
		}
	}

}
