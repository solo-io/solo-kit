// Code generated by solo-kit. DO NOT EDIT.

//Source: pkg/code-generator/codegen/templates/test_suite_template.go
package v1

import (
	"context"
	"errors"
	"fmt"

	discovery "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2"
	core "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/client"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/server"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/types"
)

// Type Definitions:

const MockXdsResourceConfigType = types.TypePrefix + "/testing.solo.io.MockXdsResourceConfig"

/* Defined a resource - to be used by snapshot */
type MockXdsResourceConfigXdsResourceWrapper struct {
	// TODO(yuval-k): This is public for mitchellh hashstructure to work properly. consider better alternatives.
	Resource *MockXdsResourceConfig
}

// Make sure the Resource interface is implemented
var _ cache.Resource = &MockXdsResourceConfigXdsResourceWrapper{}

func NewMockXdsResourceConfigXdsResourceWrapper(resourceProto *MockXdsResourceConfig) *MockXdsResourceConfigXdsResourceWrapper {
	return &MockXdsResourceConfigXdsResourceWrapper{
		Resource: resourceProto,
	}
}

func (e *MockXdsResourceConfigXdsResourceWrapper) Self() cache.XdsResourceReference {
	return cache.XdsResourceReference{Name: e.Resource.Domain, Type: MockXdsResourceConfigType}
}

func (e *MockXdsResourceConfigXdsResourceWrapper) ResourceProto() cache.ResourceProto {
	return e.Resource
}
func (e *MockXdsResourceConfigXdsResourceWrapper) References() []cache.XdsResourceReference {
	return nil
}

// Define a type record. This is used by the generic client library.
var MockXdsResourceConfigTypeRecord = client.NewTypeRecord(
	MockXdsResourceConfigType,

	// Return an empty message, that can be used to deserialize bytes into it.
	func() cache.ResourceProto { return &MockXdsResourceConfig{} },

	// Covert the message to a resource suitable for use for protobuf's Any.
	func(r cache.ResourceProto) cache.Resource {
		return &MockXdsResourceConfigXdsResourceWrapper{Resource: r.(*MockXdsResourceConfig)}
	},
)

// Server Implementation:

// Wrap the generic server and implement the type sepcific methods:
type mockXdsResourceDiscoveryServiceServer struct {
	server.Server
}

func NewMockXdsResourceDiscoveryServiceServer(genericServer server.Server) MockXdsResourceDiscoveryServiceServer {
	return &mockXdsResourceDiscoveryServiceServer{Server: genericServer}
}

func (s *mockXdsResourceDiscoveryServiceServer) StreamMockXdsResourceConfig(stream MockXdsResourceDiscoveryService_StreamMockXdsResourceConfigServer) error {
	return s.Server.StreamSolo(stream, MockXdsResourceConfigType)
}

func (s *mockXdsResourceDiscoveryServiceServer) FetchMockXdsResourceConfig(ctx context.Context, req *discovery.DiscoveryRequest) (*discovery.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = MockXdsResourceConfigType
	return s.Server.FetchSolo(ctx, req)
}

func (s *mockXdsResourceDiscoveryServiceServer) DeltaMockXdsResourceConfig(_ MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfigServer) error {
	return errors.New("not implemented")
}

// Client Implementation: Generate a strongly typed client over the generic client

// The apply functions receives resources and returns an error if they were applied correctly.
// In theory the configuration can become valid in the future (i.e. eventually consistent), but I don't think we need to worry about that now
// As our current use cases only have one configuration resource, so no interactions are expected.
type ApplyMockXdsResourceConfig func(version string, resources []*MockXdsResourceConfig) error

// Convert the strongly typed apply to a generic apply.
func applyMockXdsResourceConfig(typedApply ApplyMockXdsResourceConfig) func(cache.Resources) error {
	return func(resources cache.Resources) error {

		var configs []*MockXdsResourceConfig
		for _, r := range resources.Items {
			if proto, ok := r.ResourceProto().(*MockXdsResourceConfig); !ok {
				return fmt.Errorf("resource %s of type %s incorrect", r.Self().Name, r.Self().Type)
			} else {
				configs = append(configs, proto)
			}
		}

		return typedApply(resources.Version, configs)
	}
}

func NewMockXdsResourceConfigClient(nodeinfo *core.Node, typedApply ApplyMockXdsResourceConfig) client.Client {
	return client.NewClient(nodeinfo, MockXdsResourceConfigTypeRecord, applyMockXdsResourceConfig(typedApply))
}
