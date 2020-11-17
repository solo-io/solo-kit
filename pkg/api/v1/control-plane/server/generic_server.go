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

// Package server provides an implementation of a streaming xDS server.
package server

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_discovery_v2 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
)

type StreamV3 interface {
	Send(response *envoy_service_discovery_v3.DiscoveryResponse) error
	Recv() (*envoy_service_discovery_v3.DiscoveryRequest, error)
	grpc.ServerStream
}

type StreamV2 interface {
	Send(response *envoy_api_v2.DiscoveryResponse) error
	Recv() (*envoy_api_v2.DiscoveryRequest, error)
	grpc.ServerStream
}

// Server is a collection of handlers for streaming discovery requests.
type Server interface {
	// StreamV3 is the V3 streaming method
	StreamV3(
		stream StreamV3,
		defaultTypeURL string,
	) error
	// StreamV2 is the V2 streaming method
	StreamV2(
		stream StreamV2,
		defaultTypeURL string,
	) error
	// Fetch is the universal fetch method.
	FetchV3(
		context.Context,
		*envoy_service_discovery_v3.DiscoveryRequest,
	) (*envoy_service_discovery_v3.DiscoveryResponse, error)
	FetchV2(
		context.Context,
		*envoy_api_v2.DiscoveryRequest,
	) (*envoy_api_v2.DiscoveryResponse, error)
}

// Callbacks is a collection of callbacks inserted into the server operation.
// The callbacks are invoked synchronously.
type Callbacks interface {
	// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
	OnStreamOpen(int64, string)
	// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
	OnStreamClosed(int64)
	// OnStreamRequest is called once a request is received on a stream.
	OnStreamRequest(int64, *envoy_service_discovery_v3.DiscoveryRequest)
	// OnStreamResponse is called immediately prior to sending a response on a stream.
	OnStreamResponse(int64, *envoy_service_discovery_v3.DiscoveryRequest, *envoy_service_discovery_v3.DiscoveryResponse)
	// OnFetchRequest is called for each Fetch request
	OnFetchRequest(*envoy_service_discovery_v3.DiscoveryRequest)
	// OnFetchResponse is called immediately prior to sending a response.
	OnFetchResponse(*envoy_service_discovery_v3.DiscoveryRequest, *envoy_service_discovery_v3.DiscoveryResponse)
}

// NewServer creates handlers from a config watcher and an optional logger.
func NewServer(config cache.Cache, callbacks Callbacks) Server {
	return &server{cache: config, callbacks: callbacks}
}

type server struct {
	cache     cache.Cache
	callbacks Callbacks

	// streamCount for counting bi-di streams
	streamCount int64
}

type singleWatch struct {
	resource       chan cache.Response
	resourceCancel func()
	resourceNonce  string
}

// watches for all xDS resource types
type watches struct {
	curerntwatches map[string]singleWatch
}

func newWatches() *watches {
	return &watches{
		curerntwatches: map[string]singleWatch{},
	}
}

// Cancel all watches
func (values *watches) Cancel() {
	for _, v := range values.curerntwatches {
		if v.resourceCancel != nil {
			v.resourceCancel()
		}
	}
}

func createResponse(resp *cache.Response, typeURL string) (*envoy_service_discovery_v3.DiscoveryResponse, error) {
	if resp == nil {
		return nil, errors.New("missing response")
	}
	resources := make([]*any.Any, len(resp.Resources))
	for i := 0; i < len(resp.Resources); i++ {
		data, err := proto.Marshal(resp.Resources[i].ResourceProto())
		if err != nil {
			return nil, err
		}
		resources[i] = &any.Any{
			TypeUrl: typeURL,
			Value:   data,
		}
	}
	out := &envoy_service_discovery_v3.DiscoveryResponse{
		VersionInfo: resp.Version,
		Resources:   resources,
		TypeUrl:     typeURL,
	}
	return out, nil
}

type TypedResponse struct {
	Response *cache.Response
	TypeUrl  string
}

func (s *server) StreamV3(
	stream StreamV3,
	defaultTypeURL string,
) error {
	// a channel for receiving incoming requests
	reqCh := make(chan *envoy_service_discovery_v3.DiscoveryRequest)
	reqStop := int32(0)
	go func() {
		for {
			req, err := stream.Recv()
			if atomic.LoadInt32(&reqStop) != 0 {
				return
			}
			if err != nil {
				close(reqCh)
				return
			}
			reqCh <- req
		}
	}()

	err := s.process(s.sendV3(stream), reqCh, defaultTypeURL)

	// prevents writing to a closed channel if send failed on blocked recv
	// TODO(kuat) figure out how to unblock recv through gRPC API
	atomic.StoreInt32(&reqStop, 1)

	return err
}

func (s *server) StreamV2(
	stream StreamV2,
	defaultTypeURL string,
) error {
	// a channel for receiving incoming requests
	reqCh := make(chan *envoy_service_discovery_v3.DiscoveryRequest)
	reqStop := int32(0)
	go func() {
		for {
			req, err := stream.Recv()
			if atomic.LoadInt32(&reqStop) != 0 {
				return
			}
			if err != nil {
				close(reqCh)
				return
			}
			reqCh <- util.UpgradeDiscoveryRequest(req)
		}
	}()

	err := s.process(s.sendV2(stream), reqCh, defaultTypeURL)

	// prevents writing to a closed channel if send failed on blocked recv
	// TODO(kuat) figure out how to unblock recv through gRPC API
	atomic.StoreInt32(&reqStop, 1)

	return err
}

type sendFunc func(resp cache.Response, typeURL string, streamId int64, streamNonce *int64) (string, error)

func (s *server) sendV2(
	stream envoy_service_discovery_v2.AggregatedDiscoveryService_StreamAggregatedResourcesServer,
) sendFunc {
	return func(resp cache.Response, typeURL string, streamId int64, streamNonce *int64) (string, error) {
		out, err := createResponse(&resp, typeURL)
		if err != nil {
			return "", err
		}

		// increment nonce
		*streamNonce = *streamNonce + 1
		out.Nonce = strconv.FormatInt(*streamNonce, 10)
		if s.callbacks != nil {
			s.callbacks.OnStreamResponse(streamId, &resp.Request, out)
		}
		return out.Nonce, stream.Send(util.DowngradeDiscoveryResponse(out))
	}
}

func (s *server) sendV3(
	stream envoy_service_discovery_v3.AggregatedDiscoveryService_StreamAggregatedResourcesServer,
) sendFunc {
	return func(resp cache.Response, typeURL string, streamId int64, streamNonce *int64) (string, error) {
		out, err := createResponse(&resp, typeURL)
		if err != nil {
			return "", err
		}

		// increment nonce
		*streamNonce = *streamNonce + 1
		out.Nonce = strconv.FormatInt(*streamNonce, 10)
		if s.callbacks != nil {
			s.callbacks.OnStreamResponse(streamId, &resp.Request, out)
		}
		return out.Nonce, stream.Send(out)
	}
}

// process handles a bi-di stream request
func (s *server) process(send sendFunc, reqCh <-chan *envoy_service_discovery_v3.DiscoveryRequest, defaultTypeURL string) error {
	// increment stream count
	streamID := atomic.AddInt64(&s.streamCount, 1)

	// unique nonce generator for req-resp pairs per xDS stream; the server
	// ignores stale nonces. nonce is only modified within send() function.
	var streamNonce int64

	// a collection of watches per request type
	values := newWatches()
	defer func() {
		values.Cancel()
		if s.callbacks != nil {
			s.callbacks.OnStreamClosed(streamID)
		}
	}()

	if s.callbacks != nil {
		s.callbacks.OnStreamOpen(streamID, defaultTypeURL)
	}

	responses := make(chan TypedResponse)

	// node may only be set on the first discovery request
	var node = &envoy_config_core_v3.Node{}
	for {
		select {
		// config watcher can send the requested resources types in any order
		case resp, more := <-responses:
			if !more {
				return status.Errorf(codes.Unavailable, "watching failed")
			}
			if resp.Response == nil {
				return status.Errorf(codes.Unavailable, "watching failed for "+resp.TypeUrl)
			}
			typeurl := resp.TypeUrl
			nonce, err := send(*resp.Response, typeurl, streamID, &streamNonce)
			if err != nil {
				return err
			}
			sw := values.curerntwatches[typeurl]
			sw.resourceNonce = nonce
			values.curerntwatches[typeurl] = sw

		case req, more := <-reqCh:
			// input stream ended or errored out
			if !more {
				return nil
			}
			if req == nil {
				return status.Errorf(codes.Unavailable, "empty request")
			}

			// node field in discovery request is delta-compressed
			if req.Node != nil {
				node = req.Node
			} else {
				req.Node = node
			}
			// nonces can be reused across streams; we verify nonce only if nonce is not initialized
			nonce := req.GetResponseNonce()

			// type URL is required for ADS but is implicit for xDS
			if defaultTypeURL == cache.AnyType {
				if req.TypeUrl == "" {
					return status.Errorf(codes.InvalidArgument, "type URL is required for ADS")
				}
			} else if req.TypeUrl == "" {
				req.TypeUrl = defaultTypeURL
			}

			if s.callbacks != nil {
				s.callbacks.OnStreamRequest(streamID, req)
			}

			// cancel existing watches to (re-)request a newer version
			typeurl := req.TypeUrl
			sw := values.curerntwatches[typeurl]
			if sw.resourceNonce == "" || sw.resourceNonce == nonce {
				if sw.resourceCancel != nil {
					sw.resourceCancel()
				}

				sw.resource, sw.resourceCancel = s.createWatch(responses, req)
				values.curerntwatches[typeurl] = sw
			}
		}
	}
}

func (s *server) createWatch(responses chan<- TypedResponse, req *cache.Request) (chan cache.Response, func()) {
	typeurl := req.TypeUrl

	watchedResource, cancelwatch := s.cache.CreateWatch(*req)
	var iscanceled int32
	canceled := make(chan struct{})
	cancelwrapper := func() {
		if atomic.CompareAndSwapInt32(&iscanceled, 0, 1) {
			// make sure we dont close twice
			close(canceled)
		}
		if cancelwatch != nil {
			cancelwatch()
		}
	}

	// read the watch and post things in the main channel
	go func() {

	Loop:
		for {
			select {
			case <-canceled:
				// this was canceled. goodbye
				return
			case response, ok := <-watchedResource:
				if !ok {
					// resource chan is closed. this may have happened due to cancel,
					// or due to error.
					break Loop
				}
				responses <- TypedResponse{
					Response: &response,
					TypeUrl:  typeurl,
				}
			}
		}

		if atomic.LoadInt32(&iscanceled) == 0 {
			// the cancel function was not called - this is an error
			responses <- TypedResponse{
				Response: nil,
				TypeUrl:  typeurl,
			}
		}
	}()
	return watchedResource, cancelwrapper
}

// Fetch is the universal fetch method.
func (s *server) FetchV3(
	ctx context.Context,
	req *envoy_service_discovery_v3.DiscoveryRequest,
) (*envoy_service_discovery_v3.DiscoveryResponse, error) {
	if s.callbacks != nil {
		s.callbacks.OnFetchRequest(req)
	}
	resp, err := s.cache.Fetch(ctx, *req)
	if err != nil {
		return nil, err
	}
	out, err := createResponse(resp, req.TypeUrl)
	if s.callbacks != nil {
		s.callbacks.OnFetchResponse(req, out)
	}
	return out, err
}

// Fetch is the universal fetch method.
func (s *server) FetchV2(
	ctx context.Context,
	req *envoy_api_v2.DiscoveryRequest,
) (*envoy_api_v2.DiscoveryResponse, error) {
	upgradedReq := util.UpgradeDiscoveryRequest(req)
	out, err := s.FetchV3(ctx, upgradedReq)
	return util.DowngradeDiscoveryResponse(out), err
}
