#!/usr/bin/env bash

set -e

ROOT=$(dirname "${BASH_SOURCE[0]}")/../../..
SOLO_KIT=${ROOT}/solo-kit
IN=${SOLO_KIT}/api/v1/
EXTERNAL=${SOLO_KIT}/api/external/

# code-generator does work with go.mod but makes assumptions about
# the project living in $GOPATH/src. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
#trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

IMPORTS="\
    -I=${IN} \
    -I=${EXTERNAL} \
    -I=${ROOT}
    -I=vendor/github.com/solo-io/protoc-gen-ext"

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:${TEMP_DIR}"
HASH_FLAG="--ext_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:${TEMP_DIR}"

INPUT_PROTOS="${IN}*.proto"

protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${HASH_FLAG} \
    ${INPUT_PROTOS}

cp -r  ${TEMP_DIR}/github.com/solo-io/solo-kit ${ROOT}

goimports -w pkg