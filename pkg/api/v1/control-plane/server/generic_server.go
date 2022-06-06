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

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	sk_discovery "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/util"
	solo_discovery "github.com/solo-io/solo-kit/pkg/api/xds"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
)

type StreamEnvoyV3 interface {
	Send(response *envoy_service_discovery_v3.DiscoveryResponse) error
	Recv() (*envoy_service_discovery_v3.DiscoveryRequest, error)
	grpc.ServerStream
}

type StreamSolo interface {
	Send(response *sk_discovery.DiscoveryResponse) error
	Recv() (*sk_discovery.DiscoveryRequest, error)
	grpc.ServerStream
}

// Server is a collection of handlers for streaming discovery requests.
type Server interface {
	// StreamEnvoyV3 is the streaming method for Evnoy V3 XDS
	StreamEnvoyV3(
		stream StreamEnvoyV3,
		defaultTypeURL string,
	) error
	// StreamSolo is the streaming method for Solo discovery
	StreamSolo(
		stream StreamSolo,
		defaultTypeURL string,
	) error
	// Fetch is the universal fetch method.
	FetchEnvoyV3(
		context.Context,
		*envoy_service_discovery_v3.DiscoveryRequest,
	) (*envoy_service_discovery_v3.DiscoveryResponse, error)
	FetchSolo(
		context.Context,
		*sk_discovery.DiscoveryRequest,
	) (*sk_discovery.DiscoveryResponse, error)
}

// Callbacks is a collection of callbacks inserted into the server operation.
// The callbacks are invoked synchronously.
type Callbacks interface {
	// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
	// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
	OnStreamOpen(context.Context, int64, string) error
	// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
	OnStreamClosed(int64)
	// OnStreamRequest is called once a request is received on a stream.
	// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
	OnStreamRequest(int64, *envoy_service_discovery_v3.DiscoveryRequest) error
	// OnStreamResponse is called immediately prior to sending a response on a stream.
	OnStreamResponse(int64, *envoy_service_discovery_v3.DiscoveryRequest, *envoy_service_discovery_v3.DiscoveryResponse)
	// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
	// request and respond with an error.
	OnFetchRequest(context.Context, *envoy_service_discovery_v3.DiscoveryRequest) error
	// OnFetchResponse is called immediately prior to sending a response.
	OnFetchResponse(*envoy_service_discovery_v3.DiscoveryRequest, *envoy_service_discovery_v3.DiscoveryResponse)
}

// NewServer creates handlers from a config watcher and an optional logger.
func NewServer(ctx context.Context, config cache.Cache, callbacks Callbacks) Server {
	return &server{ctx: ctx, cache: config, callbacks: callbacks}
}

// server sends requests to the cache to be fullfilled.
// This generates responses asynchronously.
// Responses are submitted back to the Envoy client or fetched from the client.
type server struct {
	// cache is an interface to handle resource cache, and to create channels when the resources are updated.
	cache cache.Cache
	// callbacks are pre and post callback used when responses and requests sent/received.
	callbacks Callbacks
	// ctx is the context in which the server is alive.
	ctx context.Context
	// streamCount for counting bi-di streams
	streamCount int64
}

// singleWatch contians a channel that can be used to watch for new responses from the cache.
type singleWatch struct {
	// resource is the channel used to receive the response.
	resource chan cache.Response
	// resourceCancel is a function that allows you to close the resource channel.
	resourceCancel func()
	// resourceNonce is the nonce used to identify the response to a request.
	resourceNonce string
}

// watches for all xDS resource types
type watches struct {
	// currentwatches are the response channels used for each resource type.
	// Currently the only purpose is to cancel the stored channels from the Cancel() function.
	currentwatches map[string]singleWatch
}

// newWatches returns an instantiation of watches
func newWatches() *watches {
	return &watches{
		currentwatches: map[string]singleWatch{},
	}
}

// Cancel all watches
func (values *watches) Cancel() {
	for _, v := range values.currentwatches {
		if v.resourceCancel != nil {
			v.resourceCancel()
		}
	}
}

// createResponse will use the response (Envoy Request with updated resources) to serialize the resources and create an Envoy Response.
//
// The response tells Envoy that the current SotW, based off the Envoy Requested resources.
// The response contains its assocaited request.
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

// TypedResponse contains the response from a xDS request with the typeURL
type TypedResponse struct {
	// Response is the response from a request
	Response *cache.Response
	// TypeUrl is the type Url of the xDS request
	TypeUrl string
}

// StreamEnvoyV3 will create a request channel that will receive requests from the streams Recv() function.
// It will then set up the processes to handle requests when they are received, so that the server can respond to the requests.
// The defaultTypeURL is used to identify the type of the resources that the Envoy stream is watching for.
//
// process is called to handle both the request and the response to the request. It does this by sending the response onto the stream.
func (s *server) StreamEnvoyV3(
	stream StreamEnvoyV3,
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

	// TODO-JAKE may have to block all requests in an order, so that the first request
	// responded to is the one that is highest on the priority list
	err := s.process(stream.Context(), s.sendEnvoyV3(stream), reqCh, defaultTypeURL)

	// prevents writing to a closed channel if send failed on blocked recv
	// TODO(kuat) figure out how to unblock recv through gRPC API
	atomic.StoreInt32(&reqStop, 1)

	return err
}

func (s *server) StreamSolo(
	stream StreamSolo,
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

	err := s.process(stream.Context(), s.sendSolo(stream), reqCh, defaultTypeURL)

	// prevents writing to a closed channel if send failed on blocked recv
	// TODO(kuat) figure out how to unblock recv through gRPC API
	atomic.StoreInt32(&reqStop, 1)

	return err
}

type sendFunc func(resp cache.Response, typeURL string, streamId int64, streamNonce *int64) (string, error)

func (s *server) sendSolo(
	stream solo_discovery.SoloDiscoveryService_StreamAggregatedResourcesServer,
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

// sendEnvoyV3 returns a function that is called to send an Envoy response. The cahe response is used to create an Envoy response and update the nonce.
//
// It will then send the response to the stream.
// This will handle any callbacks on the server as well.
func (s *server) sendEnvoyV3(
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

// process handles both the request and the response of an Envoy request.
//
// For each request submitted onto the request channel, wait for the corresponding response on the response channel.
//
// For each response received from a request submitted, the send function to send the
// response and translates it to an Envoy Response and send it back on the Envoy client.
//
// Requests are received from the servers stream.Recv() function
// Callbacks are handled as well.
func (s *server) process(
	ctx context.Context,
	send sendFunc,
	reqCh <-chan *envoy_service_discovery_v3.DiscoveryRequest,
	defaultTypeURL string,
) error {
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
		if err := s.callbacks.OnStreamOpen(ctx, streamNonce, defaultTypeURL); err != nil {
			return err
		}
	}

	responses := make(chan TypedResponse)

	// node may only be set on the first discovery request
	var node = &envoy_config_core_v3.Node{}
	for {
		select {
		case <-s.ctx.Done():
			return nil
		// config watcher can send the requested resources types in any order
		// responses come from requests submitted to createWatch()
		case resp, more := <-responses:
			if !more {
				return status.Errorf(codes.Unavailable, "watching failed")
			}
			if resp.Response == nil {
				return status.Errorf(codes.Unavailable, "watching failed for "+resp.TypeUrl)
			}
			typeurl := resp.TypeUrl
			// send the response of a request
			nonce, err := send(*resp.Response, typeurl, streamID, &streamNonce)
			if err != nil {
				return err
			}
			sw := values.currentwatches[typeurl]
			sw.resourceNonce = nonce
			values.currentwatches[typeurl] = sw

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
			if defaultTypeURL == resource.AnyType {
				if req.TypeUrl == "" {
					return status.Errorf(codes.InvalidArgument, "type URL is required for ADS")
				}
			} else if req.TypeUrl == "" {
				req.TypeUrl = defaultTypeURL
			}

			if s.callbacks != nil {
				if err := s.callbacks.OnStreamRequest(streamID, req); err != nil {
					return err
				}
			}

			// cancel existing watches to (re-)request a newer version
			typeurl := req.TypeUrl
			sw := values.currentwatches[typeurl]
			if sw.resourceNonce == "" || sw.resourceNonce == nonce {
				if sw.resourceCancel != nil {
					sw.resourceCancel()
				}

				// wait for a response on the respones channel. Send the request to generate the response asynchronously.
				sw.resource, sw.resourceCancel = s.createResponseWatch(responses, req)
				values.currentwatches[typeurl] = sw
			}
		}
	}
}

// createResponseWatch returns a channel for the response of a request and the cancel function.
// A request is used to generate an async responce that is submitted to the respones channel.
//
// It creates a go routine to send responses onto the respones channel.
// If the watch created canceled, then it will close the go routine, else there was an error and a nil response is sent.
func (s *server) createResponseWatch(responses chan<- TypedResponse, req *cache.Request) (chan cache.Response, func()) {
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
			// receive responses for the requested resources
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
func (s *server) FetchEnvoyV3(
	ctx context.Context,
	req *envoy_service_discovery_v3.DiscoveryRequest,
) (*envoy_service_discovery_v3.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	if s.callbacks != nil {
		if err := s.callbacks.OnFetchRequest(ctx, req); err != nil {
			return nil, err
		}
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
func (s *server) FetchSolo(
	ctx context.Context,
	req *sk_discovery.DiscoveryRequest,
) (*sk_discovery.DiscoveryResponse, error) {
	upgradedReq := util.UpgradeDiscoveryRequest(req)
	out, err := s.FetchEnvoyV3(ctx, upgradedReq)
	return util.DowngradeDiscoveryResponse(out), err
}
