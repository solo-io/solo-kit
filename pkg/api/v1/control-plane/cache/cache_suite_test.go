package cache_test

import (
	"fmt"
	"testing"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	v32 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/resource"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestControlPlaneCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ControlPlaneCache Suite")
}

const (
	ListenerName = "listener_0"
	UpstreamHost = "127.0.0.1"
)

var (
	// Compile-time assertion
	_ cache.Snapshot = new(TestSnapshot)
)

type TestSnapshot struct {
	// Endpoints are items in the EDS V3 response payload.
	Endpoints cache.Resources

	// Clusters are items in the CDS response payload.
	Clusters cache.Resources

	// Routes are items in the RDS response payload.
	Routes cache.Resources

	// Listeners are items in the LDS response payload.
	Listeners cache.Resources
}

func (s TestSnapshot) Consistent() error {
	endpoints := resource.GetResourceReferences(s.Clusters.Items)
	if len(endpoints) != len(s.Endpoints.Items) {
		return fmt.Errorf("mismatched endpoint reference and resource lengths: length of %v does not equal length of %v", endpoints, s.Endpoints.Items)
	}
	if err := cache.SupersetWithResource(endpoints, s.Endpoints.Items); err != nil {
		return err
	}

	routes := resource.GetResourceReferences(s.Listeners.Items)
	if len(routes) != len(s.Routes.Items) {
		return fmt.Errorf("mismatched route reference and resource lengths: length of %v does not equal length of %v", routes, s.Routes.Items)
	}
	return cache.SupersetWithResource(routes, s.Routes.Items)
}

func (s TestSnapshot) MakeConsistent() {
	endpoints := resource.GetResourceReferences(s.Clusters.Items)
	for resourceName := range s.Endpoints.Items {
		if cluster, exists := endpoints[resourceName]; !exists {
			// add placeholder
			s.Endpoints.Items[resourceName] = resource.NewEnvoyResource(
				&endpoint.ClusterLoadAssignment{
					ClusterName: cluster.Self().Name,
					Endpoints:   []*endpoint.LocalityLbEndpoints{},
				},
			)
		}
	}
	routes := resource.GetResourceReferences(s.Listeners.Items)
	for resourceName := range s.Listeners.Items {
		if listener, exists := routes[resourceName]; !exists {
			// add placeholder
			s.Routes.Items[resourceName] = resource.NewEnvoyResource(
				&route.RouteConfiguration{
					Name: fmt.Sprintf("%s-%s", listener.Self().Name, "routes-for-invalid-envoy"),
					VirtualHosts: []*route.VirtualHost{
						{
							Name:    "invalid-envoy-config-vhost",
							Domains: []string{"*"},
							Routes: []*route.Route{
								{
									Match: &route.RouteMatch{
										PathSpecifier: &route.RouteMatch_Prefix{
											Prefix: "/",
										},
									},
									Action: &route.Route_DirectResponse{
										DirectResponse: &route.DirectResponseAction{
											Status: 500,
											Body: &core.DataSource{
												Specifier: &core.DataSource_InlineString{
													InlineString: "Invalid Envoy Configuration. " +
														"This placeholder was generated to localize pain to the misconfigured route",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			)
		}
	}
}

func (s TestSnapshot) GetResources(typ string) cache.Resources {
	switch typ {
	case resource.EndpointTypeV3:
		return s.Endpoints
	case resource.ClusterTypeV3:
		return s.Clusters
	case resource.RouteTypeV3:
		return s.Routes
	case resource.ListenerTypeV3:
		return s.Listeners
	}
	return cache.Resources{}
}
func (s TestSnapshot) Clone() cache.Snapshot {
	snapshotClone := &TestSnapshot{}

	snapshotClone.Endpoints = cache.Resources{
		Version: s.Endpoints.Version,
		Items:   cloneItems(s.Endpoints.Items),
	}

	snapshotClone.Clusters = cache.Resources{
		Version: s.Clusters.Version,
		Items:   cloneItems(s.Clusters.Items),
	}

	snapshotClone.Routes = cache.Resources{
		Version: s.Routes.Version,
		Items:   cloneItems(s.Routes.Items),
	}

	snapshotClone.Listeners = cache.Resources{
		Version: s.Listeners.Version,
		Items:   cloneItems(s.Listeners.Items),
	}

	return snapshotClone
}

func cloneItems(items map[string]cache.Resource) map[string]cache.Resource {
	clonedItems := make(map[string]cache.Resource, len(items))
	for k, v := range items {
		resProto := v.ResourceProto()
		resClone := proto.Clone(resProto)
		clonedItems[k] = resource.NewEnvoyResource(resClone)
	}
	return clonedItems
}

var _ cache.Snapshot = TestSnapshot{}

func makeEndpoint(clusterName string) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  UpstreamHost,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: uint32(8080),
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func makeEDSCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       makeEndpoint(clusterName),
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &v32.ConfigSource{
				ConfigSourceSpecifier: &v32.ConfigSource_Ads{
					Ads: &v32.AggregatedConfigSource{},
				},
			},
		},
	}
}

func makeCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       makeEndpoint(clusterName),
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
	}
}

func makeRoute(routeName string, clusterName string) *route.RouteConfiguration {
	routeConfiguration := &route.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "local_service",
			Domains: []string{"*"},
		}},
	}
	routeConfiguration.VirtualHosts[0].Routes = []*route.Route{{
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterName,
				},
				HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
					HostRewriteLiteral: UpstreamHost,
				},
			},
		},
	}}
	return routeConfiguration
}

func makeHTTPListener(route string) *listener.Listener {
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    &core.ConfigSource{},
				RouteConfigName: route,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
		}},
	}
	pbst, err := ptypes.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name:    ListenerName,
		Address: &core.Address{},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}
