#!/usr/bin/env bash

set -e

set -o errexit
set -o nounset
set -o pipefail


# The following script is used to generate the solo-kit protos.
# This script will work both in and our of the GOPATH, however, it does assume that the imported protos will be
# available in the root level vendor folder. This script will be run as part of `make generated-code` so there
# should be no need to run it otherwise. `make generated-code` will also vendor the necessary protos.
ROOT=$(dirname "${BASH_SOURCE[0]}")/../../..
SOLO_KIT=${ROOT}/solo-kit
IN=${SOLO_KIT}/api/v1/
VENDOR_ROOT=vendor_any/github.com

TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

IMPORTS="\
    -I=${IN} \
    -I=${ROOT} \
    -I=${VENDOR_ROOT}/solo-io/protoc-gen-ext \
    -I=${VENDOR_ROOT}/solo-io/protoc-gen-ext/external"

GO_FLAG="--go_out=${TEMP_DIR}"
HASH_FLAG="--ext_out=${TEMP_DIR}"

INPUT_PROTOS="${IN}*.proto"

protoc ${IMPORTS} \
    ${GO_FLAG} \
    ${HASH_FLAG} \
    ${INPUT_PROTOS}

echo ">> made it to here"
cp -r  ${TEMP_DIR}/github.com/solo-io/solo-kit ${ROOT}

goimports -w pkg