
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
PACKAGE_NAME=github.com/solo-io/solo-kit
ROOT_PKG=${PACKAGE_NAME}/pkg/api/v1/clients/kube
CLIENT_PKG=${ROOT_PKG}/crd/client
APIS_PKG=${ROOT_PKG}/crd

# Below code is copied from https://github.com/weaveworks/flagger/blob/master/hack/update-codegen.sh
# Grab code-generator version from go.sum.
CODEGEN_PKG=${GOPATH}/src/k8s.io/code-generator

#if [[ ! -d ${CODEGEN_PKG} ]]; then
#    echo "${CODEGEN_PKG} is missing. Run 'go mod vendor'."
#    exit 1
#fi


echo ">> Using ${CODEGEN_PKG}"

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

# Ensure we can execute.
chmod +x ${CODEGEN_PKG}/generate-groups.sh


${CODEGEN_PKG}/generate-groups.sh all \
    ${CLIENT_PKG} \
    ${APIS_PKG} \
    solo.io:v1 \
    --output-base "${TEMP_DIR}"
# Copy everything back.
cp -r "${TEMP_DIR}/${PACKAGE_NAME}/" "${SCRIPT_ROOT}/"

