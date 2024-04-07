// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.6.1
// source: status.proto

package core

import (
	reflect "reflect"
	sync "sync"

	_struct "github.com/golang/protobuf/ptypes/struct"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Status_State int32

const (
	// Pending status indicates the resource has not yet been validated
	Status_Pending Status_State = 0
	// Accepted indicates the resource has been validated
	Status_Accepted Status_State = 1
	// Rejected indicates an invalid configuration by the user
	// Rejected resources may be propagated to the xDS server depending on their severity
	Status_Rejected Status_State = 2
	// Warning indicates a partially invalid configuration by the user
	// Resources with Warnings may be partially accepted by a controller, depending on the implementation
	Status_Warning Status_State = 3
)

// Enum value maps for Status_State.
var (
	Status_State_name = map[int32]string{
		0: "Pending",
		1: "Accepted",
		2: "Rejected",
		3: "Warning",
	}
	Status_State_value = map[string]int32{
		"Pending":  0,
		"Accepted": 1,
		"Rejected": 2,
		"Warning":  3,
	}
)

func (x Status_State) Enum() *Status_State {
	p := new(Status_State)
	*p = x
	return p
}

func (x Status_State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status_State) Descriptor() protoreflect.EnumDescriptor {
	return file_status_proto_enumTypes[0].Descriptor()
}

func (Status_State) Type() protoreflect.EnumType {
	return &file_status_proto_enumTypes[0]
}

func (x Status_State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status_State.Descriptor instead.
func (Status_State) EnumDescriptor() ([]byte, []int) {
	return file_status_proto_rawDescGZIP(), []int{1, 0}
}

// *
// NamespacedStatuses indicates the Status of a resource according to each controller.
// NamespacedStatuses are meant to be read-only by users
type NamespacedStatuses struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Mapping from namespace to the Status written by the controller running in that namespace.
	Statuses map[string]*Status `protobuf:"bytes,1,rep,name=statuses,proto3" json:"statuses,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *NamespacedStatuses) Reset() {
	*x = NamespacedStatuses{}
	if protoimpl.UnsafeEnabled {
		mi := &file_status_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NamespacedStatuses) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NamespacedStatuses) ProtoMessage() {}

func (x *NamespacedStatuses) ProtoReflect() protoreflect.Message {
	mi := &file_status_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NamespacedStatuses.ProtoReflect.Descriptor instead.
func (*NamespacedStatuses) Descriptor() ([]byte, []int) {
	return file_status_proto_rawDescGZIP(), []int{0}
}

func (x *NamespacedStatuses) GetStatuses() map[string]*Status {
	if x != nil {
		return x.Statuses
	}
	return nil
}

// *
// Status indicates whether a resource has been (in)validated by a reporter in the system.
// Statuses are meant to be read-only by users
type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// State is the enum indicating the state of the resource
	State Status_State `protobuf:"varint,1,opt,name=state,proto3,enum=core.solo.io.Status_State" json:"state,omitempty"`
	// Reason is a description of the error for Rejected resources. If the resource is pending or accepted, this field will be empty
	Reason string `protobuf:"bytes,2,opt,name=reason,proto3" json:"reason,omitempty"`
	// Reference to the reporter who wrote this status
	ReportedBy string `protobuf:"bytes,3,opt,name=reported_by,json=reportedBy,proto3" json:"reported_by,omitempty"`
	// Reference to statuses (by resource-ref string: "Kind.Namespace.Name") of subresources of the parent resource
	SubresourceStatuses map[string]*Status `protobuf:"bytes,4,rep,name=subresource_statuses,json=subresourceStatuses,proto3" json:"subresource_statuses,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Opaque details about status results
	Details *_struct.Struct `protobuf:"bytes,5,opt,name=details,proto3" json:"details,omitempty"`
	// Additional information about the current state of the resource.
	Messages []string `protobuf:"bytes,6,rep,name=Messages,proto3" json:"Messages,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_status_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_status_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_status_proto_rawDescGZIP(), []int{1}
}

func (x *Status) GetState() Status_State {
	if x != nil {
		return x.State
	}
	return Status_Pending
}

func (x *Status) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *Status) GetReportedBy() string {
	if x != nil {
		return x.ReportedBy
	}
	return ""
}

func (x *Status) GetSubresourceStatuses() map[string]*Status {
	if x != nil {
		return x.SubresourceStatuses
	}
	return nil
}

func (x *Status) GetDetails() *_struct.Struct {
	if x != nil {
		return x.Details
	}
	return nil
}

func (x *Status) GetMessages() []string {
	if x != nil {
		return x.Messages
	}
	return nil
}

type ParentReference struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Group       string `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
	Kind        string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	Namespace   string `protobuf:"bytes,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Name        string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	SectionName string `protobuf:"bytes,5,opt,name=section_name,json=sectionName,proto3" json:"section_name,omitempty"`
}

func (x *ParentReference) Reset() {
	*x = ParentReference{}
	if protoimpl.UnsafeEnabled {
		mi := &file_status_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParentReference) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParentReference) ProtoMessage() {}

func (x *ParentReference) ProtoReflect() protoreflect.Message {
	mi := &file_status_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParentReference.ProtoReflect.Descriptor instead.
func (*ParentReference) Descriptor() ([]byte, []int) {
	return file_status_proto_rawDescGZIP(), []int{2}
}

func (x *ParentReference) GetGroup() string {
	if x != nil {
		return x.Group
	}
	return ""
}

func (x *ParentReference) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *ParentReference) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *ParentReference) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ParentReference) GetSectionName() string {
	if x != nil {
		return x.SectionName
	}
	return ""
}

type KubeCondition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type               string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Status             string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	ObservedGeneration int64  `protobuf:"varint,3,opt,name=observedGeneration,proto3" json:"observedGeneration,omitempty"`
	Reason             string `protobuf:"bytes,5,opt,name=reason,proto3" json:"reason,omitempty"`
	Message            string `protobuf:"bytes,6,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *KubeCondition) Reset() {
	*x = KubeCondition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_status_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KubeCondition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KubeCondition) ProtoMessage() {}

func (x *KubeCondition) ProtoReflect() protoreflect.Message {
	mi := &file_status_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KubeCondition.ProtoReflect.Descriptor instead.
func (*KubeCondition) Descriptor() ([]byte, []int) {
	return file_status_proto_rawDescGZIP(), []int{3}
}

func (x *KubeCondition) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *KubeCondition) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *KubeCondition) GetObservedGeneration() int64 {
	if x != nil {
		return x.ObservedGeneration
	}
	return 0
}

func (x *KubeCondition) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *KubeCondition) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type PolicyAncestorStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AncestorRef    *ParentReference `protobuf:"bytes,1,opt,name=ancestor_ref,json=ancestorRef,proto3" json:"ancestor_ref,omitempty"`
	ControllerName string           `protobuf:"bytes,2,opt,name=controller_name,json=controllerName,proto3" json:"controller_name,omitempty"`
	Conditions     []*KubeCondition `protobuf:"bytes,3,rep,name=conditions,proto3" json:"conditions,omitempty"`
}

func (x *PolicyAncestorStatus) Reset() {
	*x = PolicyAncestorStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_status_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PolicyAncestorStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolicyAncestorStatus) ProtoMessage() {}

func (x *PolicyAncestorStatus) ProtoReflect() protoreflect.Message {
	mi := &file_status_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolicyAncestorStatus.ProtoReflect.Descriptor instead.
func (*PolicyAncestorStatus) Descriptor() ([]byte, []int) {
	return file_status_proto_rawDescGZIP(), []int{4}
}

func (x *PolicyAncestorStatus) GetAncestorRef() *ParentReference {
	if x != nil {
		return x.AncestorRef
	}
	return nil
}

func (x *PolicyAncestorStatus) GetControllerName() string {
	if x != nil {
		return x.ControllerName
	}
	return ""
}

func (x *PolicyAncestorStatus) GetConditions() []*KubeCondition {
	if x != nil {
		return x.Conditions
	}
	return nil
}

type PolicyStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ancestors []*PolicyAncestorStatus `protobuf:"bytes,1,rep,name=ancestors,proto3" json:"ancestors,omitempty"`
}

func (x *PolicyStatus) Reset() {
	*x = PolicyStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_status_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PolicyStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolicyStatus) ProtoMessage() {}

func (x *PolicyStatus) ProtoReflect() protoreflect.Message {
	mi := &file_status_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolicyStatus.ProtoReflect.Descriptor instead.
func (*PolicyStatus) Descriptor() ([]byte, []int) {
	return file_status_proto_rawDescGZIP(), []int{5}
}

func (x *PolicyStatus) GetAncestors() []*PolicyAncestorStatus {
	if x != nil {
		return x.Ancestors
	}
	return nil
}

var File_status_proto protoreflect.FileDescriptor

var file_status_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c,
	0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x12, 0x65, 0x78,
	0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb3,
	0x01, 0x0a, 0x12, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x65, 0x73, 0x12, 0x4a, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73,
	0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65,
	0x73, 0x1a, 0x51, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e,
	0x69, 0x6f, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0xc1, 0x03, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x30, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x70,
	0x6f, 0x72, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x42, 0x79, 0x12, 0x60, 0x0a, 0x14, 0x73, 0x75,
	0x62, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x53,
	0x75, 0x62, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x13, 0x73, 0x75, 0x62, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x12, 0x31, 0x0a, 0x07,
	0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12,
	0x1a, 0x0a, 0x08, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x08, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x1a, 0x5c, 0x0a, 0x18, 0x53,
	0x75, 0x62, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x3d, 0x0a, 0x05, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x10, 0x00, 0x12,
	0x0c, 0x0a, 0x08, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x10, 0x01, 0x12, 0x0c, 0x0a,
	0x08, 0x52, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x65, 0x64, 0x10, 0x02, 0x12, 0x0b, 0x0a, 0x07, 0x57,
	0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x10, 0x03, 0x22, 0x90, 0x01, 0x0a, 0x0f, 0x50, 0x61, 0x72,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70,
	0x61, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x9d, 0x01, 0x0a, 0x0d,
	0x4b, 0x75, 0x62, 0x65, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2e, 0x0a, 0x12, 0x6f, 0x62, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x64, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x47,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f,
	0x6e, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0xbe, 0x01, 0x0a, 0x14,
	0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x41, 0x6e, 0x63, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x40, 0x0a, 0x0c, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x74, 0x6f, 0x72,
	0x5f, 0x72, 0x65, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x6f, 0x72,
	0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x0b, 0x61, 0x6e, 0x63, 0x65, 0x73,
	0x74, 0x6f, 0x72, 0x52, 0x65, 0x66, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x3b, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e,
	0x69, 0x6f, 0x2e, 0x4b, 0x75, 0x62, 0x65, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x50, 0x0a, 0x0c,
	0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x40, 0x0a, 0x09,
	0x61, 0x6e, 0x63, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x22, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x50,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x41, 0x6e, 0x63, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x09, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x73, 0x42, 0x43,
	0xb8, 0xf5, 0x04, 0x01, 0xc0, 0xf5, 0x04, 0x01, 0xd0, 0xf5, 0x04, 0x01, 0x5a, 0x35, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f,
	0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63,
	0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_status_proto_rawDescOnce sync.Once
	file_status_proto_rawDescData = file_status_proto_rawDesc
)

func file_status_proto_rawDescGZIP() []byte {
	file_status_proto_rawDescOnce.Do(func() {
		file_status_proto_rawDescData = protoimpl.X.CompressGZIP(file_status_proto_rawDescData)
	})
	return file_status_proto_rawDescData
}

var file_status_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_status_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_status_proto_goTypes = []interface{}{
	(Status_State)(0),            // 0: core.solo.io.Status.State
	(*NamespacedStatuses)(nil),   // 1: core.solo.io.NamespacedStatuses
	(*Status)(nil),               // 2: core.solo.io.Status
	(*ParentReference)(nil),      // 3: core.solo.io.ParentReference
	(*KubeCondition)(nil),        // 4: core.solo.io.KubeCondition
	(*PolicyAncestorStatus)(nil), // 5: core.solo.io.PolicyAncestorStatus
	(*PolicyStatus)(nil),         // 6: core.solo.io.PolicyStatus
	nil,                          // 7: core.solo.io.NamespacedStatuses.StatusesEntry
	nil,                          // 8: core.solo.io.Status.SubresourceStatusesEntry
	(*_struct.Struct)(nil),       // 9: google.protobuf.Struct
}
var file_status_proto_depIdxs = []int32{
	7, // 0: core.solo.io.NamespacedStatuses.statuses:type_name -> core.solo.io.NamespacedStatuses.StatusesEntry
	0, // 1: core.solo.io.Status.state:type_name -> core.solo.io.Status.State
	8, // 2: core.solo.io.Status.subresource_statuses:type_name -> core.solo.io.Status.SubresourceStatusesEntry
	9, // 3: core.solo.io.Status.details:type_name -> google.protobuf.Struct
	3, // 4: core.solo.io.PolicyAncestorStatus.ancestor_ref:type_name -> core.solo.io.ParentReference
	4, // 5: core.solo.io.PolicyAncestorStatus.conditions:type_name -> core.solo.io.KubeCondition
	5, // 6: core.solo.io.PolicyStatus.ancestors:type_name -> core.solo.io.PolicyAncestorStatus
	2, // 7: core.solo.io.NamespacedStatuses.StatusesEntry.value:type_name -> core.solo.io.Status
	2, // 8: core.solo.io.Status.SubresourceStatusesEntry.value:type_name -> core.solo.io.Status
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_status_proto_init() }
func file_status_proto_init() {
	if File_status_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_status_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NamespacedStatuses); i {
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
		file_status_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
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
		file_status_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParentReference); i {
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
		file_status_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KubeCondition); i {
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
		file_status_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PolicyAncestorStatus); i {
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
		file_status_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PolicyStatus); i {
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
			RawDescriptor: file_status_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_status_proto_goTypes,
		DependencyIndexes: file_status_proto_depIdxs,
		EnumInfos:         file_status_proto_enumTypes,
		MessageInfos:      file_status_proto_msgTypes,
	}.Build()
	File_status_proto = out.File
	file_status_proto_rawDesc = nil
	file_status_proto_goTypes = nil
	file_status_proto_depIdxs = nil
}
