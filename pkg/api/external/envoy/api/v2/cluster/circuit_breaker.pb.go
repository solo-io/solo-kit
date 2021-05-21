// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/cluster/circuit_breaker.proto

package cluster

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	core "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/core"
	_type "github.com/solo-io/solo-kit/pkg/api/external/envoy/type"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// :ref:`Circuit breaking<arch_overview_circuit_break>` settings can be
// specified individually for each defined priority.
type CircuitBreakers struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// If multiple :ref:`Thresholds<envoy_api_msg_cluster.CircuitBreakers.Thresholds>`
	// are defined with the same :ref:`RoutingPriority<envoy_api_enum_core.RoutingPriority>`,
	// the first one in the list is used. If no Thresholds is defined for a given
	// :ref:`RoutingPriority<envoy_api_enum_core.RoutingPriority>`, the default values
	// are used.
	Thresholds []*CircuitBreakers_Thresholds `protobuf:"bytes,1,rep,name=thresholds,proto3" json:"thresholds,omitempty"`
}

func (x *CircuitBreakers) Reset() {
	*x = CircuitBreakers{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CircuitBreakers) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CircuitBreakers) ProtoMessage() {}

func (x *CircuitBreakers) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CircuitBreakers.ProtoReflect.Descriptor instead.
func (*CircuitBreakers) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescGZIP(), []int{0}
}

func (x *CircuitBreakers) GetThresholds() []*CircuitBreakers_Thresholds {
	if x != nil {
		return x.Thresholds
	}
	return nil
}

// A Thresholds defines CircuitBreaker settings for a
// :ref:`RoutingPriority<envoy_api_enum_core.RoutingPriority>`.
// [#next-free-field: 9]
type CircuitBreakers_Thresholds struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The :ref:`RoutingPriority<envoy_api_enum_core.RoutingPriority>`
	// the specified CircuitBreaker settings apply to.
	Priority core.RoutingPriority `protobuf:"varint,1,opt,name=priority,proto3,enum=solo.io.envoy.api.v2.core.RoutingPriority" json:"priority,omitempty"`
	// The maximum number of connections that Envoy will make to the upstream
	// cluster. If not specified, the default is 1024.
	MaxConnections *wrappers.UInt32Value `protobuf:"bytes,2,opt,name=max_connections,json=maxConnections,proto3" json:"max_connections,omitempty"`
	// The maximum number of pending requests that Envoy will allow to the
	// upstream cluster. If not specified, the default is 1024.
	MaxPendingRequests *wrappers.UInt32Value `protobuf:"bytes,3,opt,name=max_pending_requests,json=maxPendingRequests,proto3" json:"max_pending_requests,omitempty"`
	// The maximum number of parallel requests that Envoy will make to the
	// upstream cluster. If not specified, the default is 1024.
	MaxRequests *wrappers.UInt32Value `protobuf:"bytes,4,opt,name=max_requests,json=maxRequests,proto3" json:"max_requests,omitempty"`
	// The maximum number of parallel retries that Envoy will allow to the
	// upstream cluster. If not specified, the default is 3.
	MaxRetries *wrappers.UInt32Value `protobuf:"bytes,5,opt,name=max_retries,json=maxRetries,proto3" json:"max_retries,omitempty"`
	// Specifies a limit on concurrent retries in relation to the number of active requests. This
	// parameter is optional.
	//
	// .. note::
	//
	//    If this field is set, the retry budget will override any configured retry circuit
	//    breaker.
	RetryBudget *CircuitBreakers_Thresholds_RetryBudget `protobuf:"bytes,8,opt,name=retry_budget,json=retryBudget,proto3" json:"retry_budget,omitempty"`
	// If track_remaining is true, then stats will be published that expose
	// the number of resources remaining until the circuit breakers open. If
	// not specified, the default is false.
	//
	// .. note::
	//
	//    If a retry budget is used in lieu of the max_retries circuit breaker,
	//    the remaining retry resources remaining will not be tracked.
	TrackRemaining bool `protobuf:"varint,6,opt,name=track_remaining,json=trackRemaining,proto3" json:"track_remaining,omitempty"`
	// The maximum number of connection pools per cluster that Envoy will concurrently support at
	// once. If not specified, the default is unlimited. Set this for clusters which create a
	// large number of connection pools. See
	// :ref:`Circuit Breaking <arch_overview_circuit_break_cluster_maximum_connection_pools>` for
	// more details.
	MaxConnectionPools *wrappers.UInt32Value `protobuf:"bytes,7,opt,name=max_connection_pools,json=maxConnectionPools,proto3" json:"max_connection_pools,omitempty"`
}

func (x *CircuitBreakers_Thresholds) Reset() {
	*x = CircuitBreakers_Thresholds{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CircuitBreakers_Thresholds) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CircuitBreakers_Thresholds) ProtoMessage() {}

func (x *CircuitBreakers_Thresholds) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CircuitBreakers_Thresholds.ProtoReflect.Descriptor instead.
func (*CircuitBreakers_Thresholds) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescGZIP(), []int{0, 0}
}

func (x *CircuitBreakers_Thresholds) GetPriority() core.RoutingPriority {
	if x != nil {
		return x.Priority
	}
	return core.RoutingPriority_DEFAULT
}

func (x *CircuitBreakers_Thresholds) GetMaxConnections() *wrappers.UInt32Value {
	if x != nil {
		return x.MaxConnections
	}
	return nil
}

func (x *CircuitBreakers_Thresholds) GetMaxPendingRequests() *wrappers.UInt32Value {
	if x != nil {
		return x.MaxPendingRequests
	}
	return nil
}

func (x *CircuitBreakers_Thresholds) GetMaxRequests() *wrappers.UInt32Value {
	if x != nil {
		return x.MaxRequests
	}
	return nil
}

func (x *CircuitBreakers_Thresholds) GetMaxRetries() *wrappers.UInt32Value {
	if x != nil {
		return x.MaxRetries
	}
	return nil
}

func (x *CircuitBreakers_Thresholds) GetRetryBudget() *CircuitBreakers_Thresholds_RetryBudget {
	if x != nil {
		return x.RetryBudget
	}
	return nil
}

func (x *CircuitBreakers_Thresholds) GetTrackRemaining() bool {
	if x != nil {
		return x.TrackRemaining
	}
	return false
}

func (x *CircuitBreakers_Thresholds) GetMaxConnectionPools() *wrappers.UInt32Value {
	if x != nil {
		return x.MaxConnectionPools
	}
	return nil
}

type CircuitBreakers_Thresholds_RetryBudget struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Specifies the limit on concurrent retries as a percentage of the sum of active requests and
	// active pending requests. For example, if there are 100 active requests and the
	// budget_percent is set to 25, there may be 25 active retries.
	//
	// This parameter is optional. Defaults to 20%.
	BudgetPercent *_type.Percent `protobuf:"bytes,1,opt,name=budget_percent,json=budgetPercent,proto3" json:"budget_percent,omitempty"`
	// Specifies the minimum retry concurrency allowed for the retry budget. The limit on the
	// number of active retries may never go below this number.
	//
	// This parameter is optional. Defaults to 3.
	MinRetryConcurrency *wrappers.UInt32Value `protobuf:"bytes,2,opt,name=min_retry_concurrency,json=minRetryConcurrency,proto3" json:"min_retry_concurrency,omitempty"`
}

func (x *CircuitBreakers_Thresholds_RetryBudget) Reset() {
	*x = CircuitBreakers_Thresholds_RetryBudget{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CircuitBreakers_Thresholds_RetryBudget) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CircuitBreakers_Thresholds_RetryBudget) ProtoMessage() {}

func (x *CircuitBreakers_Thresholds_RetryBudget) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CircuitBreakers_Thresholds_RetryBudget.ProtoReflect.Descriptor instead.
func (*CircuitBreakers_Thresholds_RetryBudget) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *CircuitBreakers_Thresholds_RetryBudget) GetBudgetPercent() *_type.Percent {
	if x != nil {
		return x.BudgetPercent
	}
	return nil
}

func (x *CircuitBreakers_Thresholds_RetryBudget) GetMinRetryConcurrency() *wrappers.UInt32Value {
	if x != nil {
		return x.MinRetryConcurrency
	}
	return nil
}

var File_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto protoreflect.FileDescriptor

var file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDesc = []byte{
	0x0a, 0x53, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x32, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x2f,
	0x63, 0x69, 0x72, 0x63, 0x75, 0x69, 0x74, 0x5f, 0x62, 0x72, 0x65, 0x61, 0x6b, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x65,
	0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x32, 0x2e, 0x63, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x1a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65, 0x6e,
	0x76, 0x6f, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x32, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f,
	0x62, 0x61, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x41, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73,
	0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x78, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f,
	0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77,
	0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x65, 0x78, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xeb, 0x06, 0x0a, 0x0f, 0x43,
	0x69, 0x72, 0x63, 0x75, 0x69, 0x74, 0x42, 0x72, 0x65, 0x61, 0x6b, 0x65, 0x72, 0x73, 0x12, 0x58,
	0x0a, 0x0a, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x38, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x65, 0x6e, 0x76,
	0x6f, 0x79, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x32, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x2e, 0x43, 0x69, 0x72, 0x63, 0x75, 0x69, 0x74, 0x42, 0x72, 0x65, 0x61, 0x6b, 0x65, 0x72,
	0x73, 0x2e, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x73, 0x52, 0x0a, 0x74, 0x68,
	0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x73, 0x1a, 0xfd, 0x05, 0x0a, 0x0a, 0x54, 0x68, 0x72,
	0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x73, 0x12, 0x50, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72,
	0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2a, 0x2e, 0x73, 0x6f, 0x6c, 0x6f,
	0x2e, 0x69, 0x6f, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x32,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x67, 0x50, 0x72, 0x69,
	0x6f, 0x72, 0x69, 0x74, 0x79, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x82, 0x01, 0x02, 0x10, 0x01, 0x52,
	0x08, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x12, 0x45, 0x0a, 0x0f, 0x6d, 0x61, 0x78,
	0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x0e, 0x6d, 0x61, 0x78, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x4e, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x5f, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x5f,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x12, 0x6d, 0x61,
	0x78, 0x50, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73,
	0x12, 0x3f, 0x0a, 0x0c, 0x6d, 0x61, 0x78, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x0b, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x73, 0x12, 0x3d, 0x0a, 0x0b, 0x6d, 0x61, 0x78, 0x5f, 0x72, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x0a, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x73,
	0x12, 0x67, 0x0a, 0x0c, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x62, 0x75, 0x64, 0x67, 0x65, 0x74,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x44, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f,
	0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x32, 0x2e, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x43, 0x69, 0x72, 0x63, 0x75, 0x69, 0x74, 0x42, 0x72, 0x65,
	0x61, 0x6b, 0x65, 0x72, 0x73, 0x2e, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x73,
	0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x42, 0x75, 0x64, 0x67, 0x65, 0x74, 0x52, 0x0b, 0x72, 0x65,
	0x74, 0x72, 0x79, 0x42, 0x75, 0x64, 0x67, 0x65, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x74, 0x72, 0x61,
	0x63, 0x6b, 0x5f, 0x72, 0x65, 0x6d, 0x61, 0x69, 0x6e, 0x69, 0x6e, 0x67, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0e, 0x74, 0x72, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x6d, 0x61, 0x69, 0x6e, 0x69,
	0x6e, 0x67, 0x12, 0x4e, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x12,
	0x6d, 0x61, 0x78, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6f,
	0x6c, 0x73, 0x1a, 0xa3, 0x01, 0x0a, 0x0b, 0x52, 0x65, 0x74, 0x72, 0x79, 0x42, 0x75, 0x64, 0x67,
	0x65, 0x74, 0x12, 0x42, 0x0a, 0x0e, 0x62, 0x75, 0x64, 0x67, 0x65, 0x74, 0x5f, 0x70, 0x65, 0x72,
	0x63, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x73, 0x6f, 0x6c,
	0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x2e,
	0x50, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x52, 0x0d, 0x62, 0x75, 0x64, 0x67, 0x65, 0x74, 0x50,
	0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x12, 0x50, 0x0a, 0x15, 0x6d, 0x69, 0x6e, 0x5f, 0x72, 0x65,
	0x74, 0x72, 0x79, 0x5f, 0x63, 0x6f, 0x6e, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x13, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x43, 0x6f, 0x6e,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x42, 0x4b, 0x5a, 0x41, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73,
	0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x32, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0xb8, 0xf5, 0x04,
	0x01, 0xc0, 0xf5, 0x04, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescOnce sync.Once
	file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescData = file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDesc
)

func file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescGZIP() []byte {
	file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescOnce.Do(func() {
		file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescData)
	})
	return file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDescData
}

var file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_goTypes = []interface{}{
	(*CircuitBreakers)(nil),                        // 0: solo.io.envoy.api.v2.cluster.CircuitBreakers
	(*CircuitBreakers_Thresholds)(nil),             // 1: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds
	(*CircuitBreakers_Thresholds_RetryBudget)(nil), // 2: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.RetryBudget
	(core.RoutingPriority)(0),                      // 3: solo.io.envoy.api.v2.core.RoutingPriority
	(*wrappers.UInt32Value)(nil),                   // 4: google.protobuf.UInt32Value
	(*_type.Percent)(nil),                          // 5: solo.io.envoy.type.Percent
}
var file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_depIdxs = []int32{
	1,  // 0: solo.io.envoy.api.v2.cluster.CircuitBreakers.thresholds:type_name -> solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds
	3,  // 1: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.priority:type_name -> solo.io.envoy.api.v2.core.RoutingPriority
	4,  // 2: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.max_connections:type_name -> google.protobuf.UInt32Value
	4,  // 3: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.max_pending_requests:type_name -> google.protobuf.UInt32Value
	4,  // 4: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.max_requests:type_name -> google.protobuf.UInt32Value
	4,  // 5: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.max_retries:type_name -> google.protobuf.UInt32Value
	2,  // 6: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.retry_budget:type_name -> solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.RetryBudget
	4,  // 7: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.max_connection_pools:type_name -> google.protobuf.UInt32Value
	5,  // 8: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.RetryBudget.budget_percent:type_name -> solo.io.envoy.type.Percent
	4,  // 9: solo.io.envoy.api.v2.cluster.CircuitBreakers.Thresholds.RetryBudget.min_retry_concurrency:type_name -> google.protobuf.UInt32Value
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() {
	file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_init()
}
func file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_init() {
	if File_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CircuitBreakers); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CircuitBreakers_Thresholds); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CircuitBreakers_Thresholds_RetryBudget); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_goTypes,
		DependencyIndexes: file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_depIdxs,
		MessageInfos:      file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_msgTypes,
	}.Build()
	File_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto = out.File
	file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_rawDesc = nil
	file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_goTypes = nil
	file_github_com_solo_io_solo_kit_api_external_envoy_api_v2_cluster_circuit_breaker_proto_depIdxs = nil
}
