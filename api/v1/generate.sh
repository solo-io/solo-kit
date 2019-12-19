#!/usr/bin/env bash

set -e

ROOT=$(dirname "${BASH_SOURCE[0]}")/../../..
SOLO_KIT=${ROOT}/solo-kit
IN=${SOLO_KIT}/api/v1/
EXTERNAL=${SOLO_KIT}/api/external/

IMPORTS="\
    -I=${IN} \
    -I=${EXTERNAL} \
    -I=${ROOT}
    -I=vendor/github.com/solo-io/protoc-gen-ext
    "

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:."
HASH_FLAG="--ext_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:."

INPUT_PROTOS="${IN}*.proto"

protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${HASH_FLAG} \
    ${INPUT_PROTOS}

cp -r  ${SOLO_KIT}/github.com/solo-io/solo-kit/ ${ROOT}
rm -rf ${SOLO_KIT}/github.com

goimports -w pkg