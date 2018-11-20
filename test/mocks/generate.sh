#!/usr/bin/env bash

set -ex

ROOT=${GOPATH}/src
SOLO_KIT=${ROOT}/github.com/solo-io/solo-kit
TEST_IN=${SOLO_KIT}/test/mocks
OUT=${SOLO_KIT}/test/mocks/v1

RELATIVE_ROOT=../../..

IMPORTS="-I=${TEST_IN} \
    -I=${SOLO_KIT}/api/external \
    -I=${ROOT}"

GOGO_FLAG="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:${GOPATH}/src/"
SOLO_KIT_FLAG="--plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit --solo-kit_out=${OUT} --solo-kit_opt=${PWD}/project.json,${RELATIVE_ROOT}/doc/docs"
INPUT_PROTOS="${TEST_IN}/*.proto"

mkdir -p ${OUT}
protoc ${IMPORTS} \
    ${GOGO_FLAG} \
    ${SOLO_KIT_FLAG} \
    ${INPUT_PROTOS}
