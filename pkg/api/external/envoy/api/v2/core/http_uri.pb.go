// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/core/http_uri.proto

package core

import (
	bytes "bytes"
	fmt "fmt"
	math "math"

	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
	_ "github.com/solo-io/protoc-gen-ext/ext"
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

// Envoy external URI descriptor
type HttpUri struct {
	// The HTTP server URI. It should be a full FQDN with protocol, host and path.
	//
	// Example:
	//
	// .. code-block:: yaml
	//
	//    uri: https://www.googleapis.com/oauth2/v1/certs
	//
	Uri string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	// Specify how `uri` is to be fetched. Today, this requires an explicit
	// cluster, but in the future we may support dynamic cluster creation or
	// inline DNS resolution. See `issue
	// <https://github.com/envoyproxy/envoy/issues/1606>`_.
	//
	// Types that are valid to be assigned to HttpUpstreamType:
	//	*HttpUri_Cluster
	HttpUpstreamType isHttpUri_HttpUpstreamType `protobuf_oneof:"http_upstream_type"`
	// Sets the maximum duration in milliseconds that a response can take to arrive upon request.
	Timeout              *types.Duration `protobuf:"bytes,3,opt,name=timeout,proto3" json:"timeout,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *HttpUri) Reset()         { *m = HttpUri{} }
func (m *HttpUri) String() string { return proto.CompactTextString(m) }
func (*HttpUri) ProtoMessage()    {}
func (*HttpUri) Descriptor() ([]byte, []int) {
	return fileDescriptor_442d16d325167287, []int{0}
}
func (m *HttpUri) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HttpUri.Unmarshal(m, b)
}
func (m *HttpUri) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HttpUri.Marshal(b, m, deterministic)
}
func (m *HttpUri) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HttpUri.Merge(m, src)
}
func (m *HttpUri) XXX_Size() int {
	return xxx_messageInfo_HttpUri.Size(m)
}
func (m *HttpUri) XXX_DiscardUnknown() {
	xxx_messageInfo_HttpUri.DiscardUnknown(m)
}

var xxx_messageInfo_HttpUri proto.InternalMessageInfo

type isHttpUri_HttpUpstreamType interface {
	isHttpUri_HttpUpstreamType()
	Equal(interface{}) bool
}

type HttpUri_Cluster struct {
	Cluster string `protobuf:"bytes,2,opt,name=cluster,proto3,oneof" json:"cluster,omitempty"`
}

func (*HttpUri_Cluster) isHttpUri_HttpUpstreamType() {}

func (m *HttpUri) GetHttpUpstreamType() isHttpUri_HttpUpstreamType {
	if m != nil {
		return m.HttpUpstreamType
	}
	return nil
}

func (m *HttpUri) GetUri() string {
	if m != nil {
		return m.Uri
	}
	return ""
}

func (m *HttpUri) GetCluster() string {
	if x, ok := m.GetHttpUpstreamType().(*HttpUri_Cluster); ok {
		return x.Cluster
	}
	return ""
}

func (m *HttpUri) GetTimeout() *types.Duration {
	if m != nil {
		return m.Timeout
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*HttpUri) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*HttpUri_Cluster)(nil),
	}
}

func init() {
	proto.RegisterType((*HttpUri)(nil), "envoy.api.v2.core.HttpUri")
}

func init() {
	proto.RegisterFile("github.com/solo-io/solo-kit/api/external/envoy/api/v2/core/http_uri.proto", fileDescriptor_442d16d325167287)
}

var fileDescriptor_442d16d325167287 = []byte{
	// 333 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x51, 0x41, 0x4b, 0x33, 0x31,
	0x10, 0x6d, 0xba, 0xfd, 0xbe, 0xb5, 0xb1, 0x1e, 0x5c, 0x04, 0x6b, 0x0b, 0xb5, 0x08, 0x42, 0x2f,
	0x26, 0xb0, 0xde, 0x05, 0x17, 0x0f, 0xf5, 0x56, 0x0a, 0x5e, 0xbc, 0x94, 0xb4, 0x8d, 0xdb, 0xd0,
	0x6d, 0x27, 0xa4, 0x93, 0x65, 0xfb, 0x8f, 0x44, 0xf0, 0x2e, 0x9e, 0xfa, 0x5b, 0xbc, 0xf5, 0x1f,
	0x78, 0x94, 0xcd, 0xee, 0x5e, 0x14, 0xf4, 0x94, 0x97, 0x37, 0x6f, 0x92, 0x37, 0x6f, 0xe8, 0x7d,
	0xac, 0x70, 0x61, 0xa7, 0x6c, 0x06, 0x2b, 0xbe, 0x81, 0x04, 0xae, 0x14, 0x14, 0xe7, 0x52, 0x21,
	0x17, 0x5a, 0x71, 0x99, 0xa1, 0x34, 0x6b, 0x91, 0x70, 0xb9, 0x4e, 0x61, 0xeb, 0xa8, 0x34, 0xe4,
	0x33, 0x30, 0x92, 0x2f, 0x10, 0xf5, 0xc4, 0x1a, 0xc5, 0xb4, 0x01, 0x84, 0xe0, 0xd8, 0x29, 0x98,
	0xd0, 0x8a, 0xa5, 0x21, 0xcb, 0x15, 0x9d, 0x5e, 0x0c, 0x10, 0x27, 0x92, 0x3b, 0xc1, 0xd4, 0x3e,
	0xf1, 0xb9, 0x35, 0x02, 0x15, 0xac, 0x8b, 0x96, 0xce, 0x69, 0x2a, 0x12, 0x35, 0x17, 0x28, 0x79,
	0x05, 0xca, 0xc2, 0x49, 0x0c, 0x31, 0x38, 0xc8, 0x73, 0x54, 0xb2, 0x47, 0x32, 0xc3, 0xdc, 0x50,
	0x71, 0xbd, 0x78, 0x25, 0xd4, 0x1f, 0x22, 0xea, 0x07, 0xa3, 0x82, 0x2e, 0xf5, 0xac, 0x51, 0x6d,
	0xd2, 0x27, 0x83, 0x66, 0xd4, 0x7c, 0xdf, 0xef, 0xbc, 0x86, 0xa9, 0xf7, 0xc9, 0x38, 0x67, 0x83,
	0x4b, 0xea, 0xcf, 0x12, 0xbb, 0x41, 0x69, 0xda, 0xf5, 0x6f, 0x82, 0x61, 0x6d, 0x5c, 0xd5, 0x82,
	0x5b, 0xea, 0xa3, 0x5a, 0x49, 0xb0, 0xd8, 0xf6, 0xfa, 0x64, 0x70, 0x18, 0x9e, 0xb1, 0xc2, 0x3f,
	0xab, 0xfc, 0xb3, 0xbb, 0xd2, 0x7f, 0xd4, 0xca, 0x5f, 0xf0, 0x5f, 0x48, 0xe3, 0x80, 0x84, 0xb5,
	0x71, 0xd5, 0x17, 0x75, 0x69, 0x50, 0xa4, 0xa2, 0x37, 0x68, 0xa4, 0x58, 0x4d, 0x70, 0xab, 0x65,
	0xf0, 0xef, 0x6d, 0xbf, 0xf3, 0x48, 0x94, 0x7d, 0x46, 0xe4, 0xf9, 0xa3, 0x47, 0xe8, 0xb9, 0x02,
	0xe6, 0xc2, 0xd2, 0x06, 0xb2, 0x2d, 0xfb, 0x91, 0x5b, 0xd4, 0x2a, 0xc7, 0x1a, 0xe5, 0xff, 0x8e,
	0xc8, 0xe3, 0xcd, 0x6f, 0x5b, 0xd2, 0xcb, 0xf8, 0x8f, 0x4d, 0x4d, 0xff, 0xbb, 0x01, 0xae, 0xbf,
	0x02, 0x00, 0x00, 0xff, 0xff, 0x6f, 0x82, 0xa7, 0x99, 0xee, 0x01, 0x00, 0x00,
}

func (this *HttpUri) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*HttpUri)
	if !ok {
		that2, ok := that.(HttpUri)
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
	if this.Uri != that1.Uri {
		return false
	}
	if that1.HttpUpstreamType == nil {
		if this.HttpUpstreamType != nil {
			return false
		}
	} else if this.HttpUpstreamType == nil {
		return false
	} else if !this.HttpUpstreamType.Equal(that1.HttpUpstreamType) {
		return false
	}
	if !this.Timeout.Equal(that1.Timeout) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *HttpUri_Cluster) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*HttpUri_Cluster)
	if !ok {
		that2, ok := that.(HttpUri_Cluster)
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
	if this.Cluster != that1.Cluster {
		return false
	}
	return true
}
