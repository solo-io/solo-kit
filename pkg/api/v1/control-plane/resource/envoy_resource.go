package resource

import (
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/types"
)

type EnvoyResource struct {
	ProtoMessage cache.ResourceProto
}

var _ cache.Resource = &EnvoyResource{}

func NewEnvoyResource(r cache.ResourceProto) *EnvoyResource {
	return &EnvoyResource{ProtoMessage: r}
}

// DefaultAPIVersion is the api version
const DefaultAPIVersion = envoy_config_core_v3.ApiVersion_V3

var (
	// ResponseTypes are supported response types.
	ResponseTypes = []string{
		types.EndpointTypeV3,
		types.ClusterTypeV3,
		types.RouteTypeV3,
		types.ListenerTypeV3,
	}
)

func (e *EnvoyResource) Self() cache.XdsResourceReference {
	return cache.XdsResourceReference{
		Name: e.Name(),
		Type: e.Type(),
	}
}

// GetResourceName returns the resource name for a valid xDS response type.
func (e *EnvoyResource) Name() string {
	switch v := e.ProtoMessage.(type) {
	case *envoy_config_endpoint_v3.ClusterLoadAssignment:
		return v.GetClusterName()
	case *envoy_config_cluster_v3.Cluster:
		return v.GetName()
	case *envoy_config_route_v3.RouteConfiguration:
		return v.GetName()
	case *envoy_config_listener_v3.Listener:
		return v.GetName()
	default:
		return ""
	}
}

func (e *EnvoyResource) ResourceProto() cache.ResourceProto {
	return e.ProtoMessage
}

func (e *EnvoyResource) Type() string {
	switch e.ProtoMessage.(type) {
	case *envoy_config_endpoint_v3.ClusterLoadAssignment:
		return types.EndpointTypeV3
	case *envoy_config_cluster_v3.Cluster:
		return types.ClusterTypeV3
	case *envoy_config_route_v3.RouteConfiguration:
		return types.RouteTypeV3
	case *envoy_config_listener_v3.Listener:
		return types.ListenerTypeV3
	default:
		return ""
	}
}

func (e *EnvoyResource) References() []cache.XdsResourceReference {
	out := make(map[cache.XdsResourceReference]bool)
	res := e.ProtoMessage
	if res == nil {
		return nil
	}
	switch v := res.(type) {
	case *envoy_config_endpoint_v3.ClusterLoadAssignment:
		// no dependencies
	case *envoy_config_cluster_v3.Cluster:
		// for EDS type, use cluster name or ServiceName override
		if v.GetType() == envoy_config_cluster_v3.Cluster_EDS {
			rr := cache.XdsResourceReference{
				Type: types.EndpointTypeV3,
			}
			if v.GetEdsClusterConfig().GetServiceName() != "" {
				rr.Name = v.GetEdsClusterConfig().GetServiceName()
			} else {
				rr.Name = v.GetName()
			}
			out[rr] = true
		}
	case *envoy_config_route_v3.RouteConfiguration:
		// References to clusters in both routes (and listeners) are not included
		// in the result, because the clusters are retrieved in bulk currently,
		// and not by name.
	case *envoy_config_listener_v3.Listener:
		// extract route configuration names from HTTP connection manager
		for _, chain := range v.GetFilterChains() {
			for _, filter := range chain.GetFilters() {

				{
					config := unmarshalHcmV3(filter.GetTypedConfig())
					if config != nil {
						if rDS := config.GetRds(); rDS != nil {
							rr := cache.XdsResourceReference{
								Type: types.RouteTypeV3,
								Name: rDS.GetRouteConfigName(),
							}
							out[rr] = true
						}
						continue
					}
				}

			}
		}
	}

	var references []cache.XdsResourceReference
	for k, _ := range out {
		references = append(references, k)
	}
	return references
}

// GetResourceReferences returns the names for dependent resources (EDS cluster
// names for CDS, RDS routes names for LDS).
func GetResourceReferences(resources map[string]cache.Resource) map[string]bool {
	out := make(map[string]bool)
	for _, res := range resources {
		if res == nil {
			continue
		}
		switch v := res.ResourceProto().(type) {
		case *envoy_config_endpoint_v3.ClusterLoadAssignment:
			// no dependencies
		case *envoy_config_cluster_v3.Cluster:
			// for EDS type, use cluster name or ServiceName override
			if v.GetType() == envoy_config_cluster_v3.Cluster_EDS {
				if v.GetEdsClusterConfig().GetServiceName() != "" {
					out[v.GetEdsClusterConfig().GetServiceName()] = true
				} else {
					out[v.GetName()] = true
				}
			}
		case *envoy_config_route_v3.RouteConfiguration:
			// References to clusters in both routes (and listeners) are not included
			// in the result, because the clusters are retrieved in bulk currently,
			// and not by name.
		case *envoy_config_listener_v3.Listener:
			// extract route configuration names from HTTP connection manager
			for _, chain := range v.GetFilterChains() {
				for _, filter := range chain.GetFilters() {

					{
						config := unmarshalHcmV3(filter.GetTypedConfig())
						if config != nil {
							if rDS := config.GetRds(); rDS != nil {
								out[rDS.GetRouteConfigName()] = true
							}
							continue
						}
					}

				}
			}
		}
	}
	return out
}

func unmarshalHcmV3(any *any.Any) *hcm_v3.HttpConnectionManager {
	var result hcm_v3.HttpConnectionManager
	if !ptypes.Is(any, &result) {
		return nil
	}
	ptypes.UnmarshalAny(any, &result)
	return &result
}

// GetResourceName returns the resource name for a valid xDS response type.
func GetResourceName(res cache.ResourceProto) string {
	switch v := res.(type) {
	case *envoy_config_endpoint_v3.ClusterLoadAssignment:
		return v.GetClusterName()
	case *envoy_config_cluster_v3.Cluster:
		return v.GetName()
	case *envoy_config_route_v3.RouteConfiguration:
		return v.GetName()
	case *envoy_config_listener_v3.Listener:
		return v.GetName()
	default:
		return ""
	}
}
