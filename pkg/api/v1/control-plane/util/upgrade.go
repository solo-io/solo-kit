package util

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
)

func UpgradeDiscoveryRequest(req *envoy_api_v2.DiscoveryRequest) *envoy_service_discovery_v3.DiscoveryRequest {
	return &envoy_service_discovery_v3.DiscoveryRequest{
		VersionInfo:   req.GetVersionInfo(),
		Node:          UpgradeNode(req.GetNode()),
		ResourceNames: req.GetResourceNames(),
		TypeUrl:       req.GetTypeUrl(),
		ResponseNonce: req.GetResponseNonce(),
		ErrorDetail:   req.GetErrorDetail(),
	}
}

func UpgradeNode(node *envoy_api_v2_core.Node) *envoy_config_core_v3.Node {
	upgradedNode := &envoy_config_core_v3.Node{
		Id:                 node.GetId(),
		Cluster:            node.GetCluster(),
		Metadata:           node.GetMetadata(),
		Locality:           UpgradeLocality(node.GetLocality()),
		UserAgentName:      node.GetUserAgentName(),
		Extensions:         make([]*envoy_config_core_v3.Extension, 0, len(node.GetExtensions())),
		ClientFeatures:     node.GetClientFeatures(),
		ListeningAddresses: make([]*envoy_config_core_v3.Address, 0, len(node.GetListeningAddresses())),
	}

	for _, v := range node.GetExtensions() {
		upgradedNode.Extensions = append(upgradedNode.Extensions, UpradeExtension(v))
	}

	for _, v := range node.GetListeningAddresses() {
		upgradedNode.ListeningAddresses = append(upgradedNode.ListeningAddresses, UpgradeAddress(v))
	}

	switch typed := node.GetUserAgentVersionType().(type) {
	case *envoy_api_v2_core.Node_UserAgentVersion:
		upgradedNode.UserAgentVersionType = &envoy_config_core_v3.Node_UserAgentVersion{
			UserAgentVersion: typed.UserAgentVersion,
		}
	case *envoy_api_v2_core.Node_UserAgentBuildVersion:
		upgradedNode.UserAgentVersionType = &envoy_config_core_v3.Node_UserAgentBuildVersion{
			UserAgentBuildVersion: &envoy_config_core_v3.BuildVersion{
				Version:  UpgradeSemanticVersion(typed.UserAgentBuildVersion.GetVersion()),
				Metadata: typed.UserAgentBuildVersion.GetMetadata(),
			},
		}
	}
	return upgradedNode
}

func UpgradeAddress(addr *envoy_api_v2_core.Address) *envoy_config_core_v3.Address {
	switch typed := addr.GetAddress().(type) {
	case *envoy_api_v2_core.Address_Pipe:
		return &envoy_config_core_v3.Address{
			Address: &envoy_config_core_v3.Address_Pipe{
				Pipe: &envoy_config_core_v3.Pipe{
					Path: typed.Pipe.GetPath(),
					Mode: typed.Pipe.GetMode(),
				},
			},
		}
	case *envoy_api_v2_core.Address_SocketAddress:
		socketAddr := &envoy_config_core_v3.SocketAddress{
			Protocol: envoy_config_core_v3.SocketAddress_Protocol(
				envoy_config_core_v3.SocketAddress_Protocol_value[typed.SocketAddress.GetProtocol().String()],
			),
			Address:      typed.SocketAddress.GetAddress(),
			ResolverName: typed.SocketAddress.GetResolverName(),
			Ipv4Compat:   typed.SocketAddress.GetIpv4Compat(),
		}
		switch typed.SocketAddress.GetPortSpecifier().(type) {
		case *envoy_api_v2_core.SocketAddress_PortValue:
			socketAddr.PortSpecifier = &envoy_config_core_v3.SocketAddress_PortValue{
				PortValue: typed.SocketAddress.GetPortValue(),
			}
		case *envoy_api_v2_core.SocketAddress_NamedPort:
			socketAddr.PortSpecifier = &envoy_config_core_v3.SocketAddress_NamedPort{
				NamedPort: typed.SocketAddress.GetNamedPort(),
			}
		}
		return &envoy_config_core_v3.Address{
			Address: &envoy_config_core_v3.Address_SocketAddress{
				SocketAddress: socketAddr,
			},
		}
	}
	return nil
}

func UpradeExtension(ext *envoy_api_v2_core.Extension) *envoy_config_core_v3.Extension {
	return &envoy_config_core_v3.Extension{
		Name:           ext.GetName(),
		Category:       ext.GetCategory(),
		TypeDescriptor: ext.GetTypeDescriptor(),
		Version: &envoy_config_core_v3.BuildVersion{
			Version:  UpgradeSemanticVersion(ext.GetVersion().GetVersion()),
			Metadata: ext.GetVersion().GetMetadata(),
		},
		Disabled: ext.GetDisabled(),
	}
}

func UpgradeSemanticVersion(smv *envoy_type.SemanticVersion) *envoy_type_v3.SemanticVersion {
	return &envoy_type_v3.SemanticVersion{
		MajorNumber: smv.GetMajorNumber(),
		MinorNumber: smv.GetMinorNumber(),
		Patch:       smv.GetPatch(),
	}
}

func UpgradeLocality(locality *envoy_api_v2_core.Locality) *envoy_config_core_v3.Locality {
	return &envoy_config_core_v3.Locality{
		Region:  locality.GetRegion(),
		Zone:    locality.GetZone(),
		SubZone: locality.GetSubZone(),
	}
}
