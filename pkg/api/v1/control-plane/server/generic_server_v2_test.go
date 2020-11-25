package server_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	discovery "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	test_resource "github.com/envoyproxy/go-control-plane/pkg/test/resource/v2"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/server"
	"google.golang.org/grpc"
)

type mockStreamV2 struct {
	t         *testing.T
	ctx       context.Context
	recv      chan *discovery.DiscoveryRequest
	sent      chan *discovery.DiscoveryResponse
	nonce     int
	sendError bool
	grpc.ServerStream
}

func (stream *mockStreamV2) Context() context.Context {
	return stream.ctx
}

func (stream *mockStreamV2) Send(resp *discovery.DiscoveryResponse) error {
	// check that nonce is monotonically incrementing
	stream.nonce = stream.nonce + 1
	if resp.Nonce != fmt.Sprintf("%d", stream.nonce) {
		stream.t.Errorf("Nonce => got %q, want %d", resp.Nonce, stream.nonce)
	}
	// check that version is set
	if resp.VersionInfo == "" {
		stream.t.Error("VersionInfo => got none, want non-empty")
	}
	// check resources are non-empty
	if len(resp.Resources) == 0 {
		stream.t.Error("Resources => got none, want non-empty")
	}
	// check that type URL matches in resources
	if resp.TypeUrl == "" {
		stream.t.Error("TypeUrl => got none, want non-empty")
	}
	for _, res := range resp.Resources {
		if res.TypeUrl != resp.TypeUrl {
			stream.t.Errorf("TypeUrl => got %q, want %q", res.TypeUrl, resp.TypeUrl)
		}
	}
	stream.sent <- resp
	if stream.sendError {
		return errors.New("send error")
	}
	return nil
}

func (stream *mockStreamV2) Recv() (*discovery.DiscoveryRequest, error) {
	req, more := <-stream.recv
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func makeMockStreamV2(t *testing.T) *mockStreamV2 {
	return &mockStreamV2{
		t:    t,
		ctx:  context.Background(),
		sent: make(chan *discovery.DiscoveryResponse, 10),
		recv: make(chan *discovery.DiscoveryRequest, 10),
	}
}

var (
	nodeV2 = &core.Node{
		Id:      "test-id",
		Cluster: "test-cluster",
	}
	testTypesV2 = []string{
		resource.EndpointTypeV2,
		resource.ClusterTypeV2,
		resource.RouteTypeV2,
		resource.ListenerTypeV2,
	}
)

func makeResponses() map[string][]cache.Response {
	return map[string][]cache.Response{
		resource.EndpointTypeV2: {{
			Version: "1",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(test_resource.MakeEndpoint(clusterName, 8080)),
			},
		}},
		resource.ClusterTypeV2: {{
			Version: "2",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(test_resource.MakeCluster(test_resource.Ads, clusterName)),
			},
		}},
		resource.RouteTypeV2: {{
			Version: "3",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(test_resource.MakeRoute(routeName, clusterName)),
			},
		}},
		resource.ListenerTypeV2: {{
			Version: "4",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(
					test_resource.MakeHTTPListener(test_resource.Ads, listenerName, 80, routeName),
				),
			},
		}},
	}
}

func TestServerShutdownV2(t *testing.T) {
	for _, typ := range testTypesV2 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponses()
			shutdown := make(chan bool)
			ctx, cancel := context.WithCancel(context.Background())
			s := server.NewServer(ctx, config, &callbacks{})

			// make a request
			resp := makeMockStreamV2(t)
			resp.recv <- &discovery.DiscoveryRequest{Node: nodeV2}
			go func() {
				var err error
				switch typ {
				case resource.EndpointTypeV2:
					err = s.StreamV2(resp, resource.EndpointTypeV2)
				case resource.ClusterTypeV2:
					err = s.StreamV2(resp, resource.ClusterTypeV2)
				case resource.RouteTypeV2:
					err = s.StreamV2(resp, resource.RouteTypeV2)
				case resource.ListenerTypeV2:
					err = s.StreamV2(resp, resource.ListenerTypeV2)
				}
				if err != nil {
					t.Errorf("Stream() => got %v, want no error", err)
				}
				shutdown <- true
			}()

			go func() {
				defer cancel()
			}()

			select {
			case <-shutdown:
			case <-time.After(1 * time.Second):
				t.Fatalf("got no response")
			}
		})
	}
}

func TestResponseHandlersV2(t *testing.T) {
	for _, typ := range testTypesV2 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponses()
			s := server.NewServer(context.Background(), config, &callbacks{})

			// make a request
			resp := makeMockStreamV2(t)
			resp.recv <- &discovery.DiscoveryRequest{Node: nodeV2}
			go func() {
				var err error
				switch typ {
				case resource.EndpointTypeV2:
					err = s.StreamV2(resp, resource.EndpointTypeV2)
				case resource.ClusterTypeV2:
					err = s.StreamV2(resp, resource.ClusterTypeV2)
				case resource.RouteTypeV2:
					err = s.StreamV2(resp, resource.RouteTypeV2)
				case resource.ListenerTypeV2:
					err = s.StreamV2(resp, resource.ListenerTypeV2)
				}
				if err != nil {
					t.Errorf("Stream() => got %v, want no error", err)
				}
			}()

			// check a response
			select {
			case <-resp.sent:
				close(resp.recv)
				if want := map[string]int{typ: 1}; !reflect.DeepEqual(want, config.counts) {
					t.Errorf("watch counts => got %v, want %v", config.counts, want)
				}
			case <-time.After(1 * time.Second):
				t.Fatalf("got no response")
			}
		})
	}
}

func TestFetchV2(t *testing.T) {
	config := makeMockConfigWatcherV3()
	config.responses = makeResponses()
	cb := &callbacks{}
	s := server.NewServer(context.Background(), config, cb)
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.EndpointTypeV2,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for endpoints: %v", err)
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ClusterTypeV2,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for clusters: %v", err)
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.RouteTypeV2,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for routes: %v", err)
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ListenerTypeV2,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for listeners: %v", err)
	}

	// try again and expect empty results
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.EndpointTypeV2,
	}); out != nil {
		t.Errorf("expected empty or error for endpoints: %v", err)
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ClusterTypeV2,
	}); out != nil {
		t.Errorf("expected empty or error for clusters: %v", err)
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.RouteTypeV2,
	}); out != nil {
		t.Errorf("expected empty or error for routes: %v", err)
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ListenerTypeV2,
	}); out != nil {
		t.Errorf("expected empty or error for listeners: %v", err)
	}

	// try empty requests: not valid in a real gRPC server
	if out, err := s.FetchV2(context.Background(), nil); out != nil {
		t.Errorf("expected empty on empty request: %v", err)
	}

	// send error from callback
	cb.callbackError = true
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.EndpointTypeV2,
	}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ClusterTypeV2,
	}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.RouteTypeV2,
	}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchV2(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ListenerTypeV2,
	}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}

	// verify fetch callbacks
	if want := 8; cb.fetchReq != want {
		t.Errorf("unexpected number of fetch requests: got %d, want %d", cb.fetchReq, want)
	}
	if want := 4; cb.fetchResp != want {
		t.Errorf("unexpected number of fetch responses: got %d, want %d", cb.fetchResp, want)
	}
}

func TestWatchClosedV2(t *testing.T) {
	for _, typ := range testTypesV2 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.closeWatch = true
			s := server.NewServer(context.Background(), config, &callbacks{})

			// make a request
			resp := makeMockStreamV2(t)
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV2,
				TypeUrl: typ,
			}

			// check that response fails since watch gets closed
			if err := s.StreamV2(resp, resource.AnyType); err == nil {
				t.Error("Stream() => got no error, want watch failed")
			}

			close(resp.recv)
		})
	}
}

func TestSendErrorV2(t *testing.T) {
	for _, typ := range testTypesV2 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponses()
			s := server.NewServer(context.Background(), config, &callbacks{})

			// make a request
			resp := makeMockStreamV2(t)
			resp.sendError = true
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV2,
				TypeUrl: typ,
			}

			// check that response fails since send returns error
			if err := s.StreamV2(resp, resource.AnyType); err == nil {
				t.Error("Stream() => got no error, want send error")
			}

			close(resp.recv)
		})
	}
}

func TestStaleNonceV2(t *testing.T) {
	for _, typ := range testTypesV2 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponses()
			s := server.NewServer(context.Background(), config, &callbacks{})

			resp := makeMockStreamV2(t)
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV2,
				TypeUrl: typ,
			}
			stop := make(chan struct{})
			go func() {
				if err := s.StreamV2(resp, resource.AnyType); err != nil {
					t.Errorf("StreamAggregatedResources() => got %v, want no error", err)
				}
				// should be two watches called
				if want := map[string]int{typ: 2}; !reflect.DeepEqual(want, config.counts) {
					t.Errorf("watch counts => got %v, want %v", config.counts, want)
				}
				close(stop)
			}()
			select {
			case <-resp.sent:
				// stale request
				resp.recv <- &discovery.DiscoveryRequest{
					Node:          nodeV2,
					TypeUrl:       typ,
					ResponseNonce: "xyz",
				}
				// fresh request
				resp.recv <- &discovery.DiscoveryRequest{
					VersionInfo:   "1",
					Node:          nodeV2,
					TypeUrl:       typ,
					ResponseNonce: "1",
				}
				close(resp.recv)
			case <-time.After(1 * time.Second):
				t.Fatalf("got %d messages on the stream, not 4", resp.nonce)
			}
			<-stop
		})
	}
}

func TestAggregatedHandlersV2(t *testing.T) {
	config := makeMockConfigWatcherV3()
	config.responses = makeResponses()
	resp := makeMockStreamV2(t)

	resp.recv <- &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ListenerTypeV2,
	}
	resp.recv <- &discovery.DiscoveryRequest{
		Node:    nodeV2,
		TypeUrl: resource.ClusterTypeV2,
	}
	resp.recv <- &discovery.DiscoveryRequest{
		Node:          nodeV2,
		TypeUrl:       resource.EndpointTypeV2,
		ResourceNames: []string{clusterName},
	}
	resp.recv <- &discovery.DiscoveryRequest{
		Node:          nodeV2,
		TypeUrl:       resource.RouteTypeV2,
		ResourceNames: []string{routeName},
	}

	s := server.NewServer(context.Background(), config, &callbacks{})
	go func() {
		if err := s.StreamV2(resp, resource.AnyType); err != nil {
			t.Errorf("StreamAggregatedResources() => got %v, want no error", err)
		}
	}()

	count := 0
	for {
		select {
		case <-resp.sent:
			count++
			if count >= 4 {
				close(resp.recv)
				if want := map[string]int{
					resource.EndpointTypeV2: 1,
					resource.ClusterTypeV2:  1,
					resource.RouteTypeV2:    1,
					resource.ListenerTypeV2: 1,
				}; !reflect.DeepEqual(want, config.counts) {
					t.Errorf("watch counts => got %v, want %v", config.counts, want)
				}

				// got all messages
				return
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("got %d messages on the stream, not 4", count)
		}
	}
}

func TestAggregateRequestTypeV2(t *testing.T) {
	config := makeMockConfigWatcherV3()
	s := server.NewServer(context.Background(), config, &callbacks{})
	resp := makeMockStreamV2(t)
	resp.recv <- &discovery.DiscoveryRequest{Node: nodeV2}
	if err := s.StreamV2(resp, resource.AnyType); err == nil {
		t.Error("StreamAggregatedResources() => got nil, want an error")
	}
}

func TestCallbackErrorV2(t *testing.T) {
	for _, typ := range testTypesV2 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponses()
			s := server.NewServer(context.Background(), config, &callbacks{callbackError: true})

			// make a request
			resp := makeMockStreamV2(t)
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV2,
				TypeUrl: typ,
			}

			// check that response fails since stream open returns error
			if err := s.StreamV2(resp, resource.AnyType); err == nil {
				t.Error("Stream() => got no error, want error")
			}

			close(resp.recv)
		})
	}
}
