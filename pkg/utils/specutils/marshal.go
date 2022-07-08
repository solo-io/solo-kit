package specutils

import (
	"bytes"

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
var specUnmarshaler = &jsonpb.Unmarshaler{AllowUnknownFields: true}

// UnmarshalMapToSpecProto converts a JSON representation of a resource spec to a protocol buffer object
// ignoring any unknown fields instead of returning an error
func UnmarshalMapToSpecProto(m map[string]interface{}, into proto.Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return specUnmarshaler.Unmarshal(bytes.NewBuffer(data), into)
}
