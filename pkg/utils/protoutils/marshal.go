// ilackarms: This file contains more than just proto-utils at this point. Should be split, or
// moved to a general serialization util package

package protoutils

import (
	"bytes"
	"encoding/json"

	"github.com/solo-io/solo-kit/pkg/utils/statusutils"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"
	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"sigs.k8s.io/yaml"
)

var jsonpbMarshaler = &jsonpb.Marshaler{OrigName: false}
var jsonpbMarshalerEmitZeroValues = &jsonpb.Marshaler{OrigName: false, EmitDefaults: true}
var jsonpbMarshalerEnumsAsInts = &jsonpb.Marshaler{OrigName: false, EnumsAsInts: true}
var inputResourceUnmarshaler = statusutils.NewNamespacedStatusesUnmarshaler(UnmarshalMapToProto)

func UnmarshalBytes(data []byte, into resources.Resource) error {
	if protoInto, ok := into.(proto.Message); ok {
		return jsonpb.Unmarshal(bytes.NewBuffer(data), protoInto)
	}
	return json.Unmarshal(data, into)
}

func MarshalBytes(res resources.Resource) ([]byte, error) {
	if pb, ok := res.(proto.Message); ok {
		buf := &bytes.Buffer{}
		err := jsonpbMarshaler.Marshal(buf, pb)
		return buf.Bytes(), err
	}
	return json.Marshal(res)
}

func UnmarshalYAML(data []byte, into resources.Resource) error {
	jsn, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	if protoInto, ok := into.(proto.Message); ok {
		return jsonpb.Unmarshal(bytes.NewBuffer(jsn), protoInto)
	}
	return json.Unmarshal(data, into)
}

func MarshalYAML(res resources.Resource) ([]byte, error) {
	var jsn []byte
	if pb, ok := res.(proto.Message); ok {
		buf := &bytes.Buffer{}
		if err := jsonpbMarshaler.Marshal(buf, pb); err != nil {
			return nil, err
		}
		jsn = buf.Bytes()
	} else {
		var err error
		jsn, err = json.Marshal(res)
		if err != nil {
			return nil, err
		}
	}
	return yaml.JSONToYAML(jsn)
}

func MarshalBytesEmitZeroValues(res resources.Resource) ([]byte, error) {
	if pb, ok := res.(proto.Message); ok {
		buf := &bytes.Buffer{}
		err := jsonpbMarshalerEmitZeroValues.Marshal(buf, pb)
		return buf.Bytes(), err
	}
	return json.Marshal(res)
}

func MarshalMap(from resources.Resource) (map[string]interface{}, error) {
	data, err := MarshalBytes(from)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return m, err
}

func MarshalMapEmitZeroValues(from resources.Resource) (map[string]interface{}, error) {
	data, err := MarshalBytesEmitZeroValues(from)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return m, err
}

func MarshalMapFromProto(from proto.Message) (map[string]interface{}, error) {
	out := &bytes.Buffer{}
	if err := jsonpbMarshaler.Marshal(out, from); err != nil {
		return nil, eris.Wrap(err, "failed to marshal proto to bytes")
	}

	var m map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &m); err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal bytes to map")
	}
	return m, nil
}

func MarshalMapFromProtoWithEnumsAsInts(from proto.Message) (map[string]interface{}, error) {
	out := &bytes.Buffer{}
	if err := jsonpbMarshalerEnumsAsInts.Marshal(out, from); err != nil {
		return nil, eris.Wrap(err, "failed to marshal proto to bytes")
	}

	var m map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &m); err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal bytes to map")
	}
	return m, nil
}

func UnmarshalMap(m map[string]interface{}, into resources.Resource) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return UnmarshalBytes(data, into)
}

func UnmarshalMapToProto(m map[string]interface{}, into proto.Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return jsonpb.Unmarshal(bytes.NewBuffer(data), into)
}

func UnmarshalStruct(structuredData *structpb.Struct, into interface{}) error {
	if structuredData == nil {
		return eris.New("cannot unmarshal nil proto struct")
	}
	strData, err := jsonpbMarshaler.MarshalToString(structuredData)
	if err != nil {
		return err
	}
	data := []byte(strData)
	return json.Unmarshal(data, into)
}

// ilackarms: help come up with a better name for this please
// values in stringMap are yaml encoded or error
// used by configmap resource client
func MapStringStringToMapStringInterface(stringMap map[string]string) (map[string]interface{}, error) {
	interfaceMap := make(map[string]interface{})
	for k, strVal := range stringMap {
		var interfaceVal interface{}
		if err := yaml.Unmarshal([]byte(strVal), &interfaceVal); err != nil {
			return nil, errors.Errorf("%v cannot be parsed as yaml", strVal)
		} else {
			interfaceMap[k] = interfaceVal
		}
	}
	return interfaceMap, nil
}

// reverse of previous
func MapStringInterfaceToMapStringString(interfaceMap map[string]interface{}) (map[string]string, error) {
	stringMap := make(map[string]string)
	for k, interfaceVal := range interfaceMap {
		yml, err := yaml.Marshal(interfaceVal)
		if err != nil {
			return nil, errors.Wrapf(err, "map values must be serializable to json")
		}
		stringMap[k] = string(yml)
	}
	return stringMap, nil
}

// UnmarshalResource convert raw Kube JSON to a Solo-Kit resource
// Returns an error if unknown fields are present in the raw json
func UnmarshalResource(kubeJson []byte, resource resources.Resource) error {
	var resourceCrd v1.Resource
	if err := json.Unmarshal(kubeJson, &resourceCrd); err != nil {
		return errors.Wrapf(err, "unmarshalling from raw json")
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(resourceCrd.ObjectMeta, true))
	if withStatus, ok := resource.(resources.InputResource); ok {
		inputResourceUnmarshaler.UnmarshalStatus(resourceCrd.Status, withStatus)
	}

	if resourceCrd.Spec != nil {
		if err := UnmarshalMap(*resourceCrd.Spec, resource); err != nil {
			return errors.Wrapf(err, "parsing resource from crd spec %v in namespace %v into %T", resourceCrd.Name, resourceCrd.Namespace, resource)
		}
	}

	return nil
}
