// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

import (
	"encoding/json"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"

	api "github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type metaOnly struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resourceName=fcars
// +genclient
// +genclient:noStatus
type FrequentlyChangingAnnotationsResource struct {
	v1.TypeMeta `json:",inline"`
	// +optional
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the implementation of this definition.
	// +optional
	Spec api.FrequentlyChangingAnnotationsResource `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

func (o *FrequentlyChangingAnnotationsResource) MarshalJSON() ([]byte, error) {
	spec, err := protoutils.MarshalMap(&o.Spec)
	if err != nil {
		return nil, err
	}
	delete(spec, "metadata")
	asMap := map[string]interface{}{
		"metadata":   o.ObjectMeta,
		"apiVersion": o.TypeMeta.APIVersion,
		"kind":       o.TypeMeta.Kind,
		"spec":       spec,
	}
	return json.Marshal(asMap)
}

func (o *FrequentlyChangingAnnotationsResource) UnmarshalJSON(data []byte) error {
	var metaOnly metaOnly
	if err := json.Unmarshal(data, &metaOnly); err != nil {
		return err
	}
	var spec api.FrequentlyChangingAnnotationsResource
	if err := protoutils.UnmarshalResource(data, &spec); err != nil {
		return err
	}
	spec.Metadata = nil
	*o = FrequentlyChangingAnnotationsResource{
		ObjectMeta: metaOnly.ObjectMeta,
		TypeMeta:   metaOnly.TypeMeta,
		Spec:       spec,
	}

	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// FrequentlyChangingAnnotationsResourceList is a collection of FrequentlyChangingAnnotationsResources.
type FrequentlyChangingAnnotationsResourceList struct {
	v1.TypeMeta `json:",inline"`
	// +optional
	v1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items       []FrequentlyChangingAnnotationsResource `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resourceName=mocks
// +genclient
type MockResource struct {
	v1.TypeMeta `json:",inline"`
	// +optional
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the implementation of this definition.
	// +optional
	Spec   api.MockResource        `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status core.NamespacedStatuses `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (o *MockResource) MarshalJSON() ([]byte, error) {
	spec, err := protoutils.MarshalMap(&o.Spec)
	if err != nil {
		return nil, err
	}
	delete(spec, "metadata")
	delete(spec, "namespaced_statuses")
	asMap := map[string]interface{}{
		"metadata":   o.ObjectMeta,
		"apiVersion": o.TypeMeta.APIVersion,
		"kind":       o.TypeMeta.Kind,
		"status":     o.Status,
		"spec":       spec,
	}
	return json.Marshal(asMap)
}

func (o *MockResource) UnmarshalJSON(data []byte) error {
	var metaOnly metaOnly
	if err := json.Unmarshal(data, &metaOnly); err != nil {
		return err
	}
	var spec api.MockResource
	if err := protoutils.UnmarshalResource(data, &spec); err != nil {
		return err
	}
	spec.Metadata = nil
	*o = MockResource{
		ObjectMeta: metaOnly.ObjectMeta,
		TypeMeta:   metaOnly.TypeMeta,
		Spec:       spec,
	}
	if spec.NamespacedStatuses != nil {
		o.Status = *spec.NamespacedStatuses
		o.Spec.NamespacedStatuses = nil
	}

	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MockResourceList is a collection of MockResources.
type MockResourceList struct {
	v1.TypeMeta `json:",inline"`
	// +optional
	v1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items       []MockResource `json:"items" protobuf:"bytes,2,rep,name=items"`
}
