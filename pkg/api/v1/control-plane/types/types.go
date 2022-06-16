package types

// AnyType is used only by ADS
const (
	AnyType    = ""
	TypePrefix = "type.googleapis.com"
)

// Resource types in xDS v1.
const (
	EndpointTypeV1 = TypePrefix + "/envoy.config.endpoint.v1.ClusterLoadAssignment"
	ClusterTypeV1  = TypePrefix + "/envoy.config.cluster.v1.Cluster"
	RouteTypeV1    = TypePrefix + "/envoy.config.route.v1.RouteConfiguration"
	ListenerTypeV1 = TypePrefix + "/envoy.config.listener.v1.Listener"
	SecretTypeV1   = TypePrefix + "/envoy.extensions.transport_sockets.tls.v1.Secret"
)

// Resource types in xDS v2.
const (
	EndpointTypeV2 = TypePrefix + "/envoy.config.endpoint.v2.ClusterLoadAssignment"
	ClusterTypeV2  = TypePrefix + "/envoy.config.cluster.v2.Cluster"
	RouteTypeV2    = TypePrefix + "/envoy.config.route.v2.RouteConfiguration"
	ListenerTypeV2 = TypePrefix + "/envoy.config.listener.v2.Listener"
	SecretTypeV2   = TypePrefix + "/envoy.extensions.transport_sockets.tls.v2.Secret"
)

// Resource types in xDS v3.
const (
	EndpointTypeV3 = TypePrefix + "/envoy.config.endpoint.v3.ClusterLoadAssignment"
	ClusterTypeV3  = TypePrefix + "/envoy.config.cluster.v3.Cluster"
	RouteTypeV3    = TypePrefix + "/envoy.config.route.v3.RouteConfiguration"
	ListenerTypeV3 = TypePrefix + "/envoy.config.listener.v3.Listener"
	SecretTypeV3   = TypePrefix + "/envoy.extensions.transport_sockets.tls.v3.Secret"
)

// Fetch urls in xDS v3.
const (
	FetchEndpointsV3 = "/v3/discovery:endpoints"
	FetchClustersV3  = "/v3/discovery:clusters"
	FetchListenersV3 = "/v3/discovery:listeners"
	FetchRoutesV3    = "/v3/discovery:routes"
	FetchSecretsV3   = "/v3/discovery:secrets"
)
