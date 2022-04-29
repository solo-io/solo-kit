package server_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	resource_v3 "github.com/envoyproxy/go-control-plane/pkg/test/resource/v3"

	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/server"
)

type mockConfigWatcherV3 struct {
	counts     map[string]int
	responses  map[string][]cache.Response
	closeWatch bool
}

func (config *mockConfigWatcherV3) GetStatusInfo(s string) cache.StatusInfo {
	panic("implement me")
}

func (config *mockConfigWatcherV3) GetStatusKeys() []string {
	panic("implement me")
}

func (config *mockConfigWatcherV3) CreateWatch(req discovery.DiscoveryRequest) (chan cache.Response, func()) {
	config.counts[req.TypeUrl] = config.counts[req.TypeUrl] + 1
	out := make(chan cache.Response, 1)
	if len(config.responses[req.TypeUrl]) > 0 {
		out <- config.responses[req.TypeUrl][0]
		config.responses[req.TypeUrl] = config.responses[req.TypeUrl][1:]
	} else if config.closeWatch {
		close(out)
	}
	return out, func() {}
}

func (config *mockConfigWatcherV3) Fetch(ctx context.Context, req discovery.DiscoveryRequest) (*cache.Response, error) {
	if len(config.responses[req.TypeUrl]) > 0 {
		out := config.responses[req.TypeUrl][0]
		config.responses[req.TypeUrl] = config.responses[req.TypeUrl][1:]
		return &out, nil
	}
	return nil, errors.New("missing")
}

func makeMockConfigWatcherV3() *mockConfigWatcherV3 {
	return &mockConfigWatcherV3{
		counts: make(map[string]int),
	}
}

type callbacks struct {
	fetchReq      int
	fetchResp     int
	callbackError bool
}

func (c *callbacks) OnStreamOpen(context.Context, int64, string) error {
	if c.callbackError {
		return errors.New("stream open error")
	}
	return nil
}
func (c *callbacks) OnStreamClosed(int64)                                     {}
func (c *callbacks) OnStreamRequest(int64, *discovery.DiscoveryRequest) error { return nil }
func (c *callbacks) OnStreamResponse(
	context.Context,
	int64,
	*discovery.DiscoveryRequest,
	*discovery.DiscoveryResponse,
) {
}
func (c *callbacks) OnFetchRequest(context.Context, *discovery.DiscoveryRequest) error {
	if c.callbackError {
		return errors.New("fetch request error")
	}
	c.fetchReq++
	return nil
}
func (c *callbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
	c.fetchResp++
}

type mockStreamEnvoyV3 struct {
	t         *testing.T
	ctx       context.Context
	recv      chan *discovery.DiscoveryRequest
	sent      chan *discovery.DiscoveryResponse
	nonce     int
	sendError bool
	grpc.ServerStream
}

func (stream *mockStreamEnvoyV3) Context() context.Context {
	return stream.ctx
}

func (stream *mockStreamEnvoyV3) Send(resp *discovery.DiscoveryResponse) error {
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

func (stream *mockStreamEnvoyV3) Recv() (*discovery.DiscoveryRequest, error) {
	req, more := <-stream.recv
	if !more {
		return nil, errors.New("empty")
	}
	return req, nil
}

func makeMockStreamEnvoyV3(t *testing.T) *mockStreamEnvoyV3 {
	return &mockStreamEnvoyV3{
		t:    t,
		ctx:  context.Background(),
		sent: make(chan *discovery.DiscoveryResponse, 10),
		recv: make(chan *discovery.DiscoveryRequest, 10),
	}
}

const (
	clusterName  = "cluster0"
	routeName    = "route0"
	listenerName = "listener0"
)

var (
	nodeV3 = &core.Node{
		Id:      "test-id",
		Cluster: "test-cluster",
	}
	testTypesV3 = []string{
		resource.EndpointTypeV3,
		resource.ClusterTypeV3,
		resource.RouteTypeV3,
		resource.ListenerTypeV3,
	}
)

func makeResponsesV3() map[string][]cache.Response {
	return map[string][]cache.Response{
		resource.EndpointTypeV3: {{
			Version: "1",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(resource_v3.MakeEndpoint(clusterName, 8080)),
			},
		}},
		resource.ClusterTypeV3: {{
			Version: "2",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(resource_v3.MakeCluster(resource_v3.Ads, clusterName)),
			},
		}},
		resource.RouteTypeV3: {{
			Version: "3",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(resource_v3.MakeRoute(routeName, clusterName)),
			},
		}},
		resource.ListenerTypeV3: {{
			Version: "4",
			Resources: []cache.Resource{
				resource.NewEnvoyResource(
					resource_v3.MakeHTTPListener(resource_v3.Ads, listenerName, 80, routeName),
				),
			},
		}},
	}
}

func TestServerShutdownV3(t *testing.T) {
	for _, typ := range testTypesV3 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponsesV3()
			shutdown := make(chan bool)
			ctx, cancel := context.WithCancel(context.Background())
			s := server.NewServer(ctx, config, &callbacks{})

			// make a request
			resp := makeMockStreamEnvoyV3(t)
			resp.recv <- &discovery.DiscoveryRequest{Node: nodeV3}
			go func() {
				var err error
				switch typ {
				case resource.EndpointTypeV3:
					err = s.StreamEnvoyV3(resp, resource.EndpointTypeV3)
				case resource.ClusterTypeV3:
					err = s.StreamEnvoyV3(resp, resource.ClusterTypeV3)
				case resource.RouteTypeV3:
					err = s.StreamEnvoyV3(resp, resource.RouteTypeV3)
				case resource.ListenerTypeV3:
					err = s.StreamEnvoyV3(resp, resource.ListenerTypeV3)
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

func TestResponseHandlersV3(t *testing.T) {
	for _, typ := range testTypesV3 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponsesV3()
			s := server.NewServer(context.Background(), config, &callbacks{})

			// make a request
			resp := makeMockStreamEnvoyV3(t)
			resp.recv <- &discovery.DiscoveryRequest{Node: nodeV3}
			go func() {
				var err error
				switch typ {
				case resource.EndpointTypeV3:
					err = s.StreamEnvoyV3(resp, resource.EndpointTypeV3)
				case resource.ClusterTypeV3:
					err = s.StreamEnvoyV3(resp, resource.ClusterTypeV3)
				case resource.RouteTypeV3:
					err = s.StreamEnvoyV3(resp, resource.RouteTypeV3)
				case resource.ListenerTypeV3:
					err = s.StreamEnvoyV3(resp, resource.ListenerTypeV3)
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

func TestFetchEnvoyV3(t *testing.T) {
	config := makeMockConfigWatcherV3()
	config.responses = makeResponsesV3()
	cb := &callbacks{}
	s := server.NewServer(context.Background(), config, cb)
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.EndpointTypeV3,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for endpoints: %v", err)
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ClusterTypeV3,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for clusters: %v", err)
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.RouteTypeV3,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for routes: %v", err)
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ListenerTypeV3,
	}); out == nil || err != nil {
		t.Errorf("unexpected empty or error for listeners: %v", err)
	}

	// try again and expect empty results
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.EndpointTypeV3,
	}); out != nil {
		t.Errorf("expected empty or error for endpoints: %v", err)
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ClusterTypeV3,
	}); out != nil {
		t.Errorf("expected empty or error for clusters: %v", err)
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.RouteTypeV3,
	}); out != nil {
		t.Errorf("expected empty or error for routes: %v", err)
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ListenerTypeV3,
	}); out != nil {
		t.Errorf("expected empty or error for listeners: %v", err)
	}

	// try empty requests: not valid in a real gRPC server
	if out, err := s.FetchEnvoyV3(context.Background(), nil); out != nil {
		t.Errorf("expected empty on empty request: %v", err)
	}

	// send error from callback
	cb.callbackError = true
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.EndpointTypeV3,
	}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ClusterTypeV3,
	}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.RouteTypeV3,
	}); out != nil || err == nil {
		t.Errorf("expected empty or error due to callback error")
	}
	if out, err := s.FetchEnvoyV3(context.Background(), &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ListenerTypeV3,
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

func TestWatchClosedV3(t *testing.T) {
	for _, typ := range testTypesV3 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.closeWatch = true
			s := server.NewServer(context.Background(), config, &callbacks{})

			// make a request
			resp := makeMockStreamEnvoyV3(t)
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV3,
				TypeUrl: typ,
			}

			// check that response fails since watch gets closed
			if err := s.StreamEnvoyV3(resp, resource.AnyType); err == nil {
				t.Error("Stream() => got no error, want watch failed")
			}

			close(resp.recv)
		})
	}
}

func TestSendErrorV3(t *testing.T) {
	for _, typ := range testTypesV3 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponsesV3()
			s := server.NewServer(context.Background(), config, &callbacks{})

			// make a request
			resp := makeMockStreamEnvoyV3(t)
			resp.sendError = true
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV3,
				TypeUrl: typ,
			}

			// check that response fails since send returns error
			if err := s.StreamEnvoyV3(resp, resource.AnyType); err == nil {
				t.Error("Stream() => got no error, want send error")
			}

			close(resp.recv)
		})
	}
}

func TestStaleNonceV3(t *testing.T) {
	for _, typ := range testTypesV3 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponsesV3()
			s := server.NewServer(context.Background(), config, &callbacks{})

			resp := makeMockStreamEnvoyV3(t)
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV3,
				TypeUrl: typ,
			}
			stop := make(chan struct{})
			go func() {
				if err := s.StreamEnvoyV3(resp, resource.AnyType); err != nil {
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
					Node:          nodeV3,
					TypeUrl:       typ,
					ResponseNonce: "xyz",
				}
				// fresh request
				resp.recv <- &discovery.DiscoveryRequest{
					VersionInfo:   "1",
					Node:          nodeV3,
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

func TestAggregatedHandlersV3(t *testing.T) {
	config := makeMockConfigWatcherV3()
	config.responses = makeResponsesV3()
	resp := makeMockStreamEnvoyV3(t)

	resp.recv <- &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ListenerTypeV3,
	}
	resp.recv <- &discovery.DiscoveryRequest{
		Node:    nodeV3,
		TypeUrl: resource.ClusterTypeV3,
	}
	resp.recv <- &discovery.DiscoveryRequest{
		Node:          nodeV3,
		TypeUrl:       resource.EndpointTypeV3,
		ResourceNames: []string{clusterName},
	}
	resp.recv <- &discovery.DiscoveryRequest{
		Node:          nodeV3,
		TypeUrl:       resource.RouteTypeV3,
		ResourceNames: []string{routeName},
	}

	s := server.NewServer(context.Background(), config, &callbacks{})
	go func() {
		if err := s.StreamEnvoyV3(resp, resource.AnyType); err != nil {
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
					resource.EndpointTypeV3: 1,
					resource.ClusterTypeV3:  1,
					resource.RouteTypeV3:    1,
					resource.ListenerTypeV3: 1,
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

func TestAggregateRequestTypeV3(t *testing.T) {
	config := makeMockConfigWatcherV3()
	s := server.NewServer(context.Background(), config, &callbacks{})
	resp := makeMockStreamEnvoyV3(t)
	resp.recv <- &discovery.DiscoveryRequest{Node: nodeV3}
	if err := s.StreamEnvoyV3(resp, resource.AnyType); err == nil {
		t.Error("StreamAggregatedResources() => got nil, want an error")
	}
}

func TestCallbackErrorV3(t *testing.T) {
	for _, typ := range testTypesV3 {
		t.Run(typ, func(t *testing.T) {
			config := makeMockConfigWatcherV3()
			config.responses = makeResponsesV3()
			s := server.NewServer(context.Background(), config, &callbacks{callbackError: true})

			// make a request
			resp := makeMockStreamEnvoyV3(t)
			resp.recv <- &discovery.DiscoveryRequest{
				Node:    nodeV3,
				TypeUrl: typ,
			}

			// check that response fails since stream open returns error
			if err := s.StreamEnvoyV3(resp, resource.AnyType); err == nil {
				t.Error("Stream() => got no error, want error")
			}

			close(resp.recv)
		})
	}
}
