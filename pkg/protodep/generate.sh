#!/usr/bin/env bash

set -e

set -o errexit
set -o nounset
set -o pipefail


IN=$(dirname "${BASH_SOURCE[0]}")
ROOT=$(go env GOMOD | rev | cut -c8- | rev)
VALIDATE=${ROOT}/vendor/github.com/envoyproxy/protoc-gen-validate

# code-generator does work with go.mod but makes assumptions about
# the project living in $GOPATH/src. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

IMPORTS="\
    -I=${IN} \
    -I=${VALIDATE}"

INPUT_PROTOS="api/*.proto"

GO_FLAG="--go_out=${TEMP_DIR}"
VALIDATE_FLAG="--validate_out=lang=go:${TEMP_DIR}"


protoc ${IMPORTS} \
    ${GO_FLAG} \
    ${VALIDATE_FLAG} \
    ${INPUT_PROTOS}

cp -r  ${TEMP_DIR}/github.com/solo-io/solo-kit/* ${ROOT}

goimports -w .