// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/discovery.proto

package v2

import (
	bytes "bytes"
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	core "github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/core"
	status "google.golang.org/genproto/googleapis/rpc/status"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// A DiscoveryRequest requests a set of versioned resources of the same type for
// a given Envoy node on some API.
type DiscoveryRequest struct {
	// The version_info provided in the request messages will be the version_info
	// received with the most recent successfully processed response or empty on
	// the first request. It is expected that no new request is sent after a
	// response is received until the Envoy instance is ready to ACK/NACK the new
	// configuration. ACK/NACK takes place by returning the new API config version
	// as applied or the previous API config version respectively. Each type_url
	// (see below) has an independent version associated with it.
	VersionInfo string `protobuf:"bytes,1,opt,name=version_info,json=versionInfo,proto3" json:"version_info,omitempty"`
	// The node making the request.
	Node *core.Node `protobuf:"bytes,2,opt,name=node,proto3" json:"node,omitempty"`
	// List of resources to subscribe to, e.g. list of cluster names or a route
	// configuration name. If this is empty, all resources for the API are
	// returned. LDS/CDS may have empty resource_names, which will cause all
	// resources for the Envoy instance to be returned. The LDS and CDS responses
	// will then imply a number of resources that need to be fetched via EDS/RDS,
	// which will be explicitly enumerated in resource_names.
	ResourceNames []string `protobuf:"bytes,3,rep,name=resource_names,json=resourceNames,proto3" json:"resource_names,omitempty"`
	// Type of the resource that is being requested, e.g.
	// "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment". This is implicit
	// in requests made via singleton xDS APIs such as CDS, LDS, etc. but is
	// required for ADS.
	TypeUrl string `protobuf:"bytes,4,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	// nonce corresponding to DiscoveryResponse being ACK/NACKed. See above
	// discussion on version_info and the DiscoveryResponse nonce comment. This
	// may be empty only if 1) this is a non-persistent-stream xDS such as HTTP,
	// or 2) the client has not yet accepted an update in this xDS stream (unlike
	// delta, where it is populated only for new explicit ACKs).
	ResponseNonce string `protobuf:"bytes,5,opt,name=response_nonce,json=responseNonce,proto3" json:"response_nonce,omitempty"`
	// This is populated when the previous :ref:`DiscoveryResponse <envoy_api_msg_DiscoveryResponse>`
	// failed to update configuration. The *message* field in *error_details* provides the Envoy
	// internal exception related to the failure. It is only intended for consumption during manual
	// debugging, the string provided is not guaranteed to be stable across Envoy versions.
	ErrorDetail          *status.Status `protobuf:"bytes,6,opt,name=error_detail,json=errorDetail,proto3" json:"error_detail,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *DiscoveryRequest) Reset()         { *m = DiscoveryRequest{} }
func (m *DiscoveryRequest) String() string { return proto.CompactTextString(m) }
func (*DiscoveryRequest) ProtoMessage()    {}
func (*DiscoveryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4a78570f62e6bc5c, []int{0}
}
func (m *DiscoveryRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DiscoveryRequest.Unmarshal(m, b)
}
func (m *DiscoveryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DiscoveryRequest.Marshal(b, m, deterministic)
}
func (m *DiscoveryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DiscoveryRequest.Merge(m, src)
}
func (m *DiscoveryRequest) XXX_Size() int {
	return xxx_messageInfo_DiscoveryRequest.Size(m)
}
func (m *DiscoveryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DiscoveryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DiscoveryRequest proto.InternalMessageInfo

func (m *DiscoveryRequest) GetVersionInfo() string {
	if m != nil {
		return m.VersionInfo
	}
	return ""
}

func (m *DiscoveryRequest) GetNode() *core.Node {
	if m != nil {
		return m.Node
	}
	return nil
}

func (m *DiscoveryRequest) GetResourceNames() []string {
	if m != nil {
		return m.ResourceNames
	}
	return nil
}

func (m *DiscoveryRequest) GetTypeUrl() string {
	if m != nil {
		return m.TypeUrl
	}
	return ""
}

func (m *DiscoveryRequest) GetResponseNonce() string {
	if m != nil {
		return m.ResponseNonce
	}
	return ""
}

func (m *DiscoveryRequest) GetErrorDetail() *status.Status {
	if m != nil {
		return m.ErrorDetail
	}
	return nil
}

type DiscoveryResponse struct {
	// The version of the response data.
	VersionInfo string `protobuf:"bytes,1,opt,name=version_info,json=versionInfo,proto3" json:"version_info,omitempty"`
	// The response resources. These resources are typed and depend on the API being called.
	Resources []*types.Any `protobuf:"bytes,2,rep,name=resources,proto3" json:"resources,omitempty"`
	// [#not-implemented-hide:]
	// Canary is used to support two Envoy command line flags:
	//
	// * --terminate-on-canary-transition-failure. When set, Envoy is able to
	//   terminate if it detects that configuration is stuck at canary. Consider
	//   this example sequence of updates:
	//   - Management server applies a canary config successfully.
	//   - Management server rolls back to a production config.
	//   - Envoy rejects the new production config.
	//   Since there is no sensible way to continue receiving configuration
	//   updates, Envoy will then terminate and apply production config from a
	//   clean slate.
	// * --dry-run-canary. When set, a canary response will never be applied, only
	//   validated via a dry run.
	Canary bool `protobuf:"varint,3,opt,name=canary,proto3" json:"canary,omitempty"`
	// Type URL for resources. Identifies the xDS API when muxing over ADS.
	// Must be consistent with the type_url in the 'resources' repeated Any (if non-empty).
	TypeUrl string `protobuf:"bytes,4,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	// For gRPC based subscriptions, the nonce provides a way to explicitly ack a
	// specific DiscoveryResponse in a following DiscoveryRequest. Additional
	// messages may have been sent by Envoy to the management server for the
	// previous version on the stream prior to this DiscoveryResponse, that were
	// unprocessed at response send time. The nonce allows the management server
	// to ignore any further DiscoveryRequests for the previous version until a
	// DiscoveryRequest bearing the nonce. The nonce is optional and is not
	// required for non-stream based xDS implementations.
	Nonce string `protobuf:"bytes,5,opt,name=nonce,proto3" json:"nonce,omitempty"`
	// [#not-implemented-hide:]
	// The control plane instance that sent the response.
	ControlPlane         *core.ControlPlane `protobuf:"bytes,6,opt,name=control_plane,json=controlPlane,proto3" json:"control_plane,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *DiscoveryResponse) Reset()         { *m = DiscoveryResponse{} }
func (m *DiscoveryResponse) String() string { return proto.CompactTextString(m) }
func (*DiscoveryResponse) ProtoMessage()    {}
func (*DiscoveryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4a78570f62e6bc5c, []int{1}
}
func (m *DiscoveryResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DiscoveryResponse.Unmarshal(m, b)
}
func (m *DiscoveryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DiscoveryResponse.Marshal(b, m, deterministic)
}
func (m *DiscoveryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DiscoveryResponse.Merge(m, src)
}
func (m *DiscoveryResponse) XXX_Size() int {
	return xxx_messageInfo_DiscoveryResponse.Size(m)
}
func (m *DiscoveryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DiscoveryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DiscoveryResponse proto.InternalMessageInfo

func (m *DiscoveryResponse) GetVersionInfo() string {
	if m != nil {
		return m.VersionInfo
	}
	return ""
}

func (m *DiscoveryResponse) GetResources() []*types.Any {
	if m != nil {
		return m.Resources
	}
	return nil
}

func (m *DiscoveryResponse) GetCanary() bool {
	if m != nil {
		return m.Canary
	}
	return false
}

func (m *DiscoveryResponse) GetTypeUrl() string {
	if m != nil {
		return m.TypeUrl
	}
	return ""
}

func (m *DiscoveryResponse) GetNonce() string {
	if m != nil {
		return m.Nonce
	}
	return ""
}

func (m *DiscoveryResponse) GetControlPlane() *core.ControlPlane {
	if m != nil {
		return m.ControlPlane
	}
	return nil
}

// DeltaDiscoveryRequest and DeltaDiscoveryResponse are used in a new gRPC
// endpoint for Delta xDS.
//
// With Delta xDS, the DeltaDiscoveryResponses do not need to include a full
// snapshot of the tracked resources. Instead, DeltaDiscoveryResponses are a
// diff to the state of a xDS client.
// In Delta XDS there are per-resource versions, which allow tracking state at
// the resource granularity.
// An xDS Delta session is always in the context of a gRPC bidirectional
// stream. This allows the xDS server to keep track of the state of xDS clients
// connected to it.
//
// In Delta xDS the nonce field is required and used to pair
// DeltaDiscoveryResponse to a DeltaDiscoveryRequest ACK or NACK.
// Optionally, a response message level system_version_info is present for
// debugging purposes only.
//
// DeltaDiscoveryRequest plays two independent roles. Any DeltaDiscoveryRequest
// can be either or both of: [1] informing the server of what resources the
// client has gained/lost interest in (using resource_names_subscribe and
// resource_names_unsubscribe), or [2] (N)ACKing an earlier resource update from
// the server (using response_nonce, with presence of error_detail making it a NACK).
// Additionally, the first message (for a given type_url) of a reconnected gRPC stream
// has a third role: informing the server of the resources (and their versions)
// that the client already possesses, using the initial_resource_versions field.
//
// As with state-of-the-world, when multiple resource types are multiplexed (ADS),
// all requests/acknowledgments/updates are logically walled off by type_url:
// a Cluster ACK exists in a completely separate world from a prior Route NACK.
// In particular, initial_resource_versions being sent at the "start" of every
// gRPC stream actually entails a message for each type_url, each with its own
// initial_resource_versions.
type DeltaDiscoveryRequest struct {
	// The node making the request.
	Node *core.Node `protobuf:"bytes,1,opt,name=node,proto3" json:"node,omitempty"`
	// Type of the resource that is being requested, e.g.
	// "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment".
	TypeUrl string `protobuf:"bytes,2,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	// DeltaDiscoveryRequests allow the client to add or remove individual
	// resources to the set of tracked resources in the context of a stream.
	// All resource names in the resource_names_subscribe list are added to the
	// set of tracked resources and all resource names in the resource_names_unsubscribe
	// list are removed from the set of tracked resources.
	//
	// *Unlike* state-of-the-world xDS, an empty resource_names_subscribe or
	// resource_names_unsubscribe list simply means that no resources are to be
	// added or removed to the resource list.
	// *Like* state-of-the-world xDS, the server must send updates for all tracked
	// resources, but can also send updates for resources the client has not subscribed to.
	//
	// NOTE: the server must respond with all resources listed in resource_names_subscribe,
	// even if it believes the client has the most recent version of them. The reason:
	// the client may have dropped them, but then regained interest before it had a chance
	// to send the unsubscribe message. See DeltaSubscriptionStateTest.RemoveThenAdd.
	//
	// These two fields can be set in any DeltaDiscoveryRequest, including ACKs
	// and initial_resource_versions.
	//
	// A list of Resource names to add to the list of tracked resources.
	ResourceNamesSubscribe []string `protobuf:"bytes,3,rep,name=resource_names_subscribe,json=resourceNamesSubscribe,proto3" json:"resource_names_subscribe,omitempty"`
	// A list of Resource names to remove from the list of tracked resources.
	ResourceNamesUnsubscribe []string `protobuf:"bytes,4,rep,name=resource_names_unsubscribe,json=resourceNamesUnsubscribe,proto3" json:"resource_names_unsubscribe,omitempty"`
	// Informs the server of the versions of the resources the xDS client knows of, to enable the
	// client to continue the same logical xDS session even in the face of gRPC stream reconnection.
	// It will not be populated: [1] in the very first stream of a session, since the client will
	// not yet have any resources,  [2] in any message after the first in a stream (for a given
	// type_url), since the server will already be correctly tracking the client's state.
	// (In ADS, the first message *of each type_url* of a reconnected stream populates this map.)
	// The map's keys are names of xDS resources known to the xDS client.
	// The map's values are opaque resource versions.
	InitialResourceVersions map[string]string `protobuf:"bytes,5,rep,name=initial_resource_versions,json=initialResourceVersions,proto3" json:"initial_resource_versions,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// When the DeltaDiscoveryRequest is a ACK or NACK message in response
	// to a previous DeltaDiscoveryResponse, the response_nonce must be the
	// nonce in the DeltaDiscoveryResponse.
	// Otherwise (unlike in DiscoveryRequest) response_nonce must be omitted.
	ResponseNonce string `protobuf:"bytes,6,opt,name=response_nonce,json=responseNonce,proto3" json:"response_nonce,omitempty"`
	// This is populated when the previous :ref:`DiscoveryResponse <envoy_api_msg_DiscoveryResponse>`
	// failed to update configuration. The *message* field in *error_details*
	// provides the Envoy internal exception related to the failure.
	ErrorDetail          *status.Status `protobuf:"bytes,7,opt,name=error_detail,json=errorDetail,proto3" json:"error_detail,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *DeltaDiscoveryRequest) Reset()         { *m = DeltaDiscoveryRequest{} }
func (m *DeltaDiscoveryRequest) String() string { return proto.CompactTextString(m) }
func (*DeltaDiscoveryRequest) ProtoMessage()    {}
func (*DeltaDiscoveryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4a78570f62e6bc5c, []int{2}
}
func (m *DeltaDiscoveryRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeltaDiscoveryRequest.Unmarshal(m, b)
}
func (m *DeltaDiscoveryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeltaDiscoveryRequest.Marshal(b, m, deterministic)
}
func (m *DeltaDiscoveryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeltaDiscoveryRequest.Merge(m, src)
}
func (m *DeltaDiscoveryRequest) XXX_Size() int {
	return xxx_messageInfo_DeltaDiscoveryRequest.Size(m)
}
func (m *DeltaDiscoveryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeltaDiscoveryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeltaDiscoveryRequest proto.InternalMessageInfo

func (m *DeltaDiscoveryRequest) GetNode() *core.Node {
	if m != nil {
		return m.Node
	}
	return nil
}

func (m *DeltaDiscoveryRequest) GetTypeUrl() string {
	if m != nil {
		return m.TypeUrl
	}
	return ""
}

func (m *DeltaDiscoveryRequest) GetResourceNamesSubscribe() []string {
	if m != nil {
		return m.ResourceNamesSubscribe
	}
	return nil
}

func (m *DeltaDiscoveryRequest) GetResourceNamesUnsubscribe() []string {
	if m != nil {
		return m.ResourceNamesUnsubscribe
	}
	return nil
}

func (m *DeltaDiscoveryRequest) GetInitialResourceVersions() map[string]string {
	if m != nil {
		return m.InitialResourceVersions
	}
	return nil
}

func (m *DeltaDiscoveryRequest) GetResponseNonce() string {
	if m != nil {
		return m.ResponseNonce
	}
	return ""
}

func (m *DeltaDiscoveryRequest) GetErrorDetail() *status.Status {
	if m != nil {
		return m.ErrorDetail
	}
	return nil
}

type DeltaDiscoveryResponse struct {
	// The version of the response data (used for debugging).
	SystemVersionInfo string `protobuf:"bytes,1,opt,name=system_version_info,json=systemVersionInfo,proto3" json:"system_version_info,omitempty"`
	// The response resources. These are typed resources, whose types must match
	// the type_url field.
	Resources []*Resource `protobuf:"bytes,2,rep,name=resources,proto3" json:"resources,omitempty"`
	// Type URL for resources. Identifies the xDS API when muxing over ADS.
	// Must be consistent with the type_url in the Any within 'resources' if 'resources' is non-empty.
	TypeUrl string `protobuf:"bytes,4,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	// Resources names of resources that have be deleted and to be removed from the xDS Client.
	// Removed resources for missing resources can be ignored.
	RemovedResources []string `protobuf:"bytes,6,rep,name=removed_resources,json=removedResources,proto3" json:"removed_resources,omitempty"`
	// The nonce provides a way for DeltaDiscoveryRequests to uniquely
	// reference a DeltaDiscoveryResponse when (N)ACKing. The nonce is required.
	Nonce                string   `protobuf:"bytes,5,opt,name=nonce,proto3" json:"nonce,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeltaDiscoveryResponse) Reset()         { *m = DeltaDiscoveryResponse{} }
func (m *DeltaDiscoveryResponse) String() string { return proto.CompactTextString(m) }
func (*DeltaDiscoveryResponse) ProtoMessage()    {}
func (*DeltaDiscoveryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4a78570f62e6bc5c, []int{3}
}
func (m *DeltaDiscoveryResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeltaDiscoveryResponse.Unmarshal(m, b)
}
func (m *DeltaDiscoveryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeltaDiscoveryResponse.Marshal(b, m, deterministic)
}
func (m *DeltaDiscoveryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeltaDiscoveryResponse.Merge(m, src)
}
func (m *DeltaDiscoveryResponse) XXX_Size() int {
	return xxx_messageInfo_DeltaDiscoveryResponse.Size(m)
}
func (m *DeltaDiscoveryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeltaDiscoveryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeltaDiscoveryResponse proto.InternalMessageInfo

func (m *DeltaDiscoveryResponse) GetSystemVersionInfo() string {
	if m != nil {
		return m.SystemVersionInfo
	}
	return ""
}

func (m *DeltaDiscoveryResponse) GetResources() []*Resource {
	if m != nil {
		return m.Resources
	}
	return nil
}

func (m *DeltaDiscoveryResponse) GetTypeUrl() string {
	if m != nil {
		return m.TypeUrl
	}
	return ""
}

func (m *DeltaDiscoveryResponse) GetRemovedResources() []string {
	if m != nil {
		return m.RemovedResources
	}
	return nil
}

func (m *DeltaDiscoveryResponse) GetNonce() string {
	if m != nil {
		return m.Nonce
	}
	return ""
}

type Resource struct {
	// The resource's name, to distinguish it from others of the same type of resource.
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// [#not-implemented-hide:]
	// The aliases are a list of other names that this resource can go by.
	Aliases []string `protobuf:"bytes,4,rep,name=aliases,proto3" json:"aliases,omitempty"`
	// The resource level version. It allows xDS to track the state of individual
	// resources.
	Version string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	// The resource being tracked.
	Resource             *types.Any `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Resource) Reset()         { *m = Resource{} }
func (m *Resource) String() string { return proto.CompactTextString(m) }
func (*Resource) ProtoMessage()    {}
func (*Resource) Descriptor() ([]byte, []int) {
	return fileDescriptor_4a78570f62e6bc5c, []int{4}
}
func (m *Resource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Resource.Unmarshal(m, b)
}
func (m *Resource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Resource.Marshal(b, m, deterministic)
}
func (m *Resource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Resource.Merge(m, src)
}
func (m *Resource) XXX_Size() int {
	return xxx_messageInfo_Resource.Size(m)
}
func (m *Resource) XXX_DiscardUnknown() {
	xxx_messageInfo_Resource.DiscardUnknown(m)
}

var xxx_messageInfo_Resource proto.InternalMessageInfo

func (m *Resource) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Resource) GetAliases() []string {
	if m != nil {
		return m.Aliases
	}
	return nil
}

func (m *Resource) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *Resource) GetResource() *types.Any {
	if m != nil {
		return m.Resource
	}
	return nil
}

func init() {
	proto.RegisterType((*DiscoveryRequest)(nil), "envoy.api.v2.DiscoveryRequest")
	proto.RegisterType((*DiscoveryResponse)(nil), "envoy.api.v2.DiscoveryResponse")
	proto.RegisterType((*DeltaDiscoveryRequest)(nil), "envoy.api.v2.DeltaDiscoveryRequest")
	proto.RegisterMapType((map[string]string)(nil), "envoy.api.v2.DeltaDiscoveryRequest.InitialResourceVersionsEntry")
	proto.RegisterType((*DeltaDiscoveryResponse)(nil), "envoy.api.v2.DeltaDiscoveryResponse")
	proto.RegisterType((*Resource)(nil), "envoy.api.v2.Resource")
}

func init() {
	proto.RegisterFile("github.com/solo-io/solo-kit/api/external/envoy/api/v2/discovery.proto", fileDescriptor_4a78570f62e6bc5c)
}

var fileDescriptor_4a78570f62e6bc5c = []byte{
	// 738 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x55, 0xc1, 0x6e, 0xdb, 0x46,
	0x10, 0x05, 0x25, 0x59, 0x96, 0x56, 0xb2, 0x61, 0x6f, 0x5d, 0x99, 0x16, 0x0c, 0x57, 0x15, 0x50,
	0x40, 0x80, 0x51, 0xb2, 0x50, 0x5b, 0xc0, 0x2d, 0x7a, 0x68, 0x5d, 0xf9, 0xe0, 0x1c, 0x0c, 0x83,
	0x86, 0x7d, 0xc8, 0x85, 0x58, 0x51, 0x63, 0x65, 0x61, 0x6a, 0x97, 0xd9, 0x5d, 0x0a, 0x22, 0x90,
	0x53, 0x90, 0x8f, 0xc9, 0x27, 0xe4, 0x53, 0x72, 0xc8, 0xc9, 0xff, 0x90, 0x43, 0x4e, 0x09, 0xb8,
	0x5c, 0x4a, 0xa2, 0xad, 0x08, 0x3a, 0x69, 0x67, 0xe6, 0xed, 0x70, 0x66, 0xde, 0x1b, 0x2d, 0xba,
	0x18, 0x53, 0xf5, 0x2a, 0x1e, 0x3a, 0x01, 0x9f, 0xb8, 0x92, 0x87, 0xfc, 0x57, 0xca, 0xb3, 0xdf,
	0x07, 0xaa, 0x5c, 0x12, 0x51, 0x17, 0x66, 0x0a, 0x04, 0x23, 0xa1, 0x0b, 0x6c, 0xca, 0x13, 0xed,
	0x9a, 0xf6, 0xdd, 0x11, 0x95, 0x01, 0x9f, 0x82, 0x48, 0x9c, 0x48, 0x70, 0xc5, 0x71, 0x53, 0x47,
	0x1d, 0x12, 0x51, 0x67, 0xda, 0x6f, 0x1f, 0x17, 0xb0, 0x01, 0x17, 0xe0, 0x0e, 0x89, 0x84, 0x0c,
	0xdb, 0x3e, 0x1a, 0x73, 0x3e, 0x0e, 0xc1, 0xd5, 0xd6, 0x30, 0xbe, 0x77, 0x09, 0x33, 0x69, 0xda,
	0x87, 0x26, 0x24, 0xa2, 0xc0, 0x95, 0x8a, 0xa8, 0x58, 0x9a, 0xc0, 0xc1, 0x98, 0x8f, 0xb9, 0x3e,
	0xba, 0xe9, 0xc9, 0x78, 0x31, 0xcc, 0x54, 0xe6, 0x84, 0x99, 0xca, 0x7c, 0xdd, 0xb7, 0x25, 0xb4,
	0x37, 0xc8, 0xab, 0xf3, 0xe0, 0x75, 0x0c, 0x52, 0xe1, 0x9f, 0x51, 0x73, 0x0a, 0x42, 0x52, 0xce,
	0x7c, 0xca, 0xee, 0xb9, 0x6d, 0x75, 0xac, 0x5e, 0xdd, 0x6b, 0x18, 0xdf, 0x25, 0xbb, 0xe7, 0xf8,
	0x14, 0x55, 0x18, 0x1f, 0x81, 0x5d, 0xea, 0x58, 0xbd, 0x46, 0xff, 0xd0, 0x59, 0x6e, 0xc8, 0x49,
	0x5b, 0x70, 0xae, 0xf8, 0x08, 0x3c, 0x0d, 0xc2, 0xbf, 0xa0, 0x5d, 0x01, 0x92, 0xc7, 0x22, 0x00,
	0x9f, 0x91, 0x09, 0x48, 0xbb, 0xdc, 0x29, 0xf7, 0xea, 0xde, 0x4e, 0xee, 0xbd, 0x4a, 0x9d, 0xf8,
	0x08, 0xd5, 0x54, 0x12, 0x81, 0x1f, 0x8b, 0xd0, 0xae, 0xe8, 0x4f, 0x6e, 0xa7, 0xf6, 0xad, 0x08,
	0x4d, 0x86, 0x88, 0x33, 0x09, 0x3e, 0xe3, 0x2c, 0x00, 0x7b, 0x4b, 0x03, 0x76, 0x72, 0xef, 0x55,
	0xea, 0xc4, 0x7f, 0xa2, 0x26, 0x08, 0xc1, 0x85, 0x3f, 0x02, 0x45, 0x68, 0x68, 0x57, 0x75, 0x75,
	0xd8, 0xc9, 0xe6, 0xe4, 0x88, 0x28, 0x70, 0x6e, 0xf4, 0x9c, 0xbc, 0x86, 0xc6, 0x0d, 0x34, 0xac,
	0xfb, 0xc5, 0x42, 0xfb, 0x4b, 0x43, 0xc8, 0x32, 0x6e, 0x32, 0x85, 0x3e, 0xaa, 0xe7, 0x2d, 0x48,
	0xbb, 0xd4, 0x29, 0xf7, 0x1a, 0xfd, 0x83, 0xfc, 0x63, 0x39, 0x5f, 0xce, 0x7f, 0x2c, 0xf1, 0x16,
	0x30, 0xdc, 0x42, 0xd5, 0x80, 0x30, 0x22, 0x12, 0xbb, 0xdc, 0xb1, 0x7a, 0x35, 0xcf, 0x58, 0xeb,
	0xba, 0x3f, 0x40, 0x5b, 0xcb, 0x4d, 0x67, 0x06, 0x1e, 0xa0, 0x9d, 0x80, 0x33, 0x25, 0x78, 0xe8,
	0x47, 0x21, 0x61, 0x60, 0xba, 0xfd, 0x69, 0x05, 0x17, 0xff, 0x67, 0xb8, 0xeb, 0x14, 0xe6, 0x35,
	0x83, 0x25, 0xab, 0xfb, 0xb5, 0x8c, 0x7e, 0x1c, 0x40, 0xa8, 0xc8, 0x33, 0x15, 0xe4, 0x14, 0x5b,
	0x9b, 0x50, 0xbc, 0x5c, 0x7d, 0xa9, 0x58, 0xfd, 0x19, 0xb2, 0x8b, 0xec, 0xfb, 0x32, 0x1e, 0xca,
	0x40, 0xd0, 0x21, 0x18, 0x1d, 0xb4, 0x0a, 0x3a, 0xb8, 0xc9, 0xa3, 0xf8, 0x1f, 0xd4, 0x7e, 0x72,
	0x33, 0x66, 0x8b, 0xbb, 0x15, 0x7d, 0xd7, 0x2e, 0xdc, 0xbd, 0x5d, 0xc4, 0xf1, 0x1b, 0x74, 0x44,
	0x19, 0x55, 0x94, 0x84, 0xfe, 0x3c, 0x8b, 0x21, 0x4f, 0xda, 0x5b, 0x9a, 0xac, 0x7f, 0x8b, 0x4d,
	0xad, 0x9c, 0x83, 0x73, 0x99, 0x25, 0xf1, 0x4c, 0x8e, 0x3b, 0x93, 0xe2, 0x82, 0x29, 0x91, 0x78,
	0x87, 0x74, 0x75, 0x74, 0x85, 0x62, 0xab, 0x9b, 0x28, 0x76, 0x7b, 0x23, 0xc5, 0xb6, 0x5f, 0xa0,
	0xe3, 0x75, 0x65, 0xe1, 0x3d, 0x54, 0x7e, 0x80, 0xc4, 0x48, 0x36, 0x3d, 0xa6, 0x1a, 0x9a, 0x92,
	0x30, 0x06, 0xc3, 0x4e, 0x66, 0xfc, 0x5d, 0x3a, 0xb3, 0xba, 0x9f, 0x2c, 0xd4, 0x7a, 0xda, 0xb9,
	0x59, 0x01, 0x07, 0xfd, 0x20, 0x13, 0xa9, 0x60, 0xe2, 0xaf, 0xd8, 0x84, 0xfd, 0x2c, 0x74, 0xb7,
	0xb4, 0x0f, 0x7f, 0x3c, 0xdf, 0x87, 0x56, 0x71, 0xc4, 0x79, 0xb9, 0xcb, 0x1b, 0xb1, 0x46, 0xf9,
	0xa7, 0x68, 0x5f, 0xc0, 0x84, 0x4f, 0x61, 0xe4, 0x2f, 0x12, 0x57, 0x35, 0xf1, 0x7b, 0x26, 0xe0,
	0xcd, 0xf3, 0xac, 0x5c, 0x93, 0xee, 0x3b, 0x0b, 0xd5, 0x72, 0x0c, 0xc6, 0xa8, 0x92, 0x0a, 0x49,
	0xaf, 0x5e, 0xdd, 0xd3, 0x67, 0x6c, 0xa3, 0x6d, 0x12, 0x52, 0x22, 0x41, 0x1a, 0x49, 0xe5, 0x66,
	0x1a, 0x31, 0x7d, 0x9b, 0x96, 0x73, 0x13, 0xff, 0x86, 0x6a, 0x79, 0x3d, 0xe6, 0x2f, 0x70, 0xf5,
	0xde, 0xcf, 0x51, 0xe7, 0xf1, 0x87, 0xcf, 0x15, 0xeb, 0xfd, 0xe3, 0x89, 0xf5, 0xf1, 0xf1, 0xc4,
	0x42, 0x6d, 0xca, 0xb3, 0xb9, 0x44, 0x82, 0xcf, 0x92, 0xc2, 0x88, 0xce, 0x77, 0xe7, 0x3c, 0x5c,
	0xa7, 0xa9, 0xae, 0xad, 0x97, 0x7f, 0xad, 0x7b, 0x75, 0xa2, 0x87, 0xf1, 0xf7, 0x5f, 0x9e, 0x61,
	0x55, 0x97, 0xf3, 0xfb, 0xb7, 0x00, 0x00, 0x00, 0xff, 0xff, 0x1c, 0xa5, 0x11, 0xbb, 0xb9, 0x06,
	0x00, 0x00,
}

func (this *DiscoveryRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DiscoveryRequest)
	if !ok {
		that2, ok := that.(DiscoveryRequest)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.VersionInfo != that1.VersionInfo {
		return false
	}
	if !this.Node.Equal(that1.Node) {
		return false
	}
	if len(this.ResourceNames) != len(that1.ResourceNames) {
		return false
	}
	for i := range this.ResourceNames {
		if this.ResourceNames[i] != that1.ResourceNames[i] {
			return false
		}
	}
	if this.TypeUrl != that1.TypeUrl {
		return false
	}
	if this.ResponseNonce != that1.ResponseNonce {
		return false
	}
	if !this.ErrorDetail.Equal(that1.ErrorDetail) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *DiscoveryResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DiscoveryResponse)
	if !ok {
		that2, ok := that.(DiscoveryResponse)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.VersionInfo != that1.VersionInfo {
		return false
	}
	if len(this.Resources) != len(that1.Resources) {
		return false
	}
	for i := range this.Resources {
		if !this.Resources[i].Equal(that1.Resources[i]) {
			return false
		}
	}
	if this.Canary != that1.Canary {
		return false
	}
	if this.TypeUrl != that1.TypeUrl {
		return false
	}
	if this.Nonce != that1.Nonce {
		return false
	}
	if !this.ControlPlane.Equal(that1.ControlPlane) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *DeltaDiscoveryRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DeltaDiscoveryRequest)
	if !ok {
		that2, ok := that.(DeltaDiscoveryRequest)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !this.Node.Equal(that1.Node) {
		return false
	}
	if this.TypeUrl != that1.TypeUrl {
		return false
	}
	if len(this.ResourceNamesSubscribe) != len(that1.ResourceNamesSubscribe) {
		return false
	}
	for i := range this.ResourceNamesSubscribe {
		if this.ResourceNamesSubscribe[i] != that1.ResourceNamesSubscribe[i] {
			return false
		}
	}
	if len(this.ResourceNamesUnsubscribe) != len(that1.ResourceNamesUnsubscribe) {
		return false
	}
	for i := range this.ResourceNamesUnsubscribe {
		if this.ResourceNamesUnsubscribe[i] != that1.ResourceNamesUnsubscribe[i] {
			return false
		}
	}
	if len(this.InitialResourceVersions) != len(that1.InitialResourceVersions) {
		return false
	}
	for i := range this.InitialResourceVersions {
		if this.InitialResourceVersions[i] != that1.InitialResourceVersions[i] {
			return false
		}
	}
	if this.ResponseNonce != that1.ResponseNonce {
		return false
	}
	if !this.ErrorDetail.Equal(that1.ErrorDetail) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *DeltaDiscoveryResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DeltaDiscoveryResponse)
	if !ok {
		that2, ok := that.(DeltaDiscoveryResponse)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.SystemVersionInfo != that1.SystemVersionInfo {
		return false
	}
	if len(this.Resources) != len(that1.Resources) {
		return false
	}
	for i := range this.Resources {
		if !this.Resources[i].Equal(that1.Resources[i]) {
			return false
		}
	}
	if this.TypeUrl != that1.TypeUrl {
		return false
	}
	if len(this.RemovedResources) != len(that1.RemovedResources) {
		return false
	}
	for i := range this.RemovedResources {
		if this.RemovedResources[i] != that1.RemovedResources[i] {
			return false
		}
	}
	if this.Nonce != that1.Nonce {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *Resource) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Resource)
	if !ok {
		that2, ok := that.(Resource)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if len(this.Aliases) != len(that1.Aliases) {
		return false
	}
	for i := range this.Aliases {
		if this.Aliases[i] != that1.Aliases[i] {
			return false
		}
	}
	if this.Version != that1.Version {
		return false
	}
	if !this.Resource.Equal(that1.Resource) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
