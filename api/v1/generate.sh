#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src
SOLO_KIT=${ROOT}/github.com/solo-io/solo-kit
IN=${SOLO_KIT}/api/v1/
EXTERNAL=${SOLO_KIT}/api/external/
OUT=${SOLO_KIT}/pkg/api/external/istio/encryption/v1/

IMPORTS="\
    -I=${IN} \
    -I=${EXTERNAL} \
    -I=${ROOT}
    -I=vendor/github.com/solo-io/protoc-gen-ext
    "

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:${GOPATH}/src/"
HASH_FLAG="--ext_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:${GOPATH}/src/"

INPUT_PROTOS="${IN}*.proto"

mkdir -p ${OUT}
protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${HASH_FLAG} \
    ${INPUT_PROTOS}