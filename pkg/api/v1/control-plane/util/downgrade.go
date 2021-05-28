package util

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_api_v2 "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2"
	envoy_api_v2_core "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/core"
)

func DowngradeDiscoveryResponse(resp *envoy_service_discovery_v3.DiscoveryResponse) *envoy_api_v2.DiscoveryResponse {
	if resp == nil {
		return nil
	}
	return &envoy_api_v2.DiscoveryResponse{
		VersionInfo:  resp.GetVersionInfo(),
		Resources:    resp.GetResources(),
		Canary:       resp.GetCanary(),
		TypeUrl:      resp.GetTypeUrl(),
		Nonce:        resp.GetNonce(),
		ControlPlane: DowngradeControlPlane(resp.GetControlPlane()),
	}
}

func DowngradeControlPlane(cp *envoy_config_core_v3.ControlPlane) *envoy_api_v2_core.ControlPlane {
	if cp == nil {
		return nil
	}
	return &envoy_api_v2_core.ControlPlane{
		Identifier: cp.GetIdentifier(),
	}
}
