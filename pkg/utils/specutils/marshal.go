package specutils

import (
	"bytes"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"

	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// specUnmarshaler is responsible for converting from a JSON representation
// of a resource spec to a protocol buffer object.
//
// We intentionally enable AllowUnknownFields when unmarshalling a resource spec
// This allows forwards compatibility in our API, where an older version of a resource
// client can read a newer version of a resource, and ignore the fields that are irrelevant.
// This scenario would arise during a rollback of an installation
var specProtoUnmarshaler = &jsonpb.Unmarshaler{AllowUnknownFields: true}

// UnmarshalSpecMapToProto converts a JSON representation of a resource spec to a protocol buffer object
// ignoring any unknown fields instead of returning an error
func UnmarshalSpecMapToProto(m map[string]interface{}, into proto.Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return specProtoUnmarshaler.Unmarshal(bytes.NewBuffer(data), into)
}

// UnmarshalSpecMapToResource converts a JSON representation of a resource spec to a Resource object
// ignoring any unknown fields instead of returning an error
func UnmarshalSpecMapToResource(m map[string]interface{}, into resources.Resource) error {
	if protoInto, ok := into.(proto.Message); ok {
		return UnmarshalSpecMapToProto(m, protoInto)
	}

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, into)
}
