package kube

import (
	"text/template"

	"github.com/solo-io/solo-kit/pkg/code-generator/codegen/templates"
)

// TODO(marco): replace hardcoded types
var GenerateScriptTemplate = template.Must(template.New("kube_generate").Funcs(templates.Funcs).Parse(`
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
ROOT_PKG={{ .ProjectConfig.GoPackage }}
CLIENT_PKG=${ROOT_PKG}/kube/client
APIS_PKG=${ROOT_PKG}/kube/apis

# Below code is copied from https://github.com/weaveworks/flagger/blob/master/hack/update-codegen.sh
CODEGEN_PKG=$(go list -f {{` + `"'{{ .Dir }}'"` + `}} -m k8s.io/code-generator)
# With k8s.io/code-generator v0.28.x the boilerplate file has been removed. So we get it from k8s.io/gengo instead
GENGO_PKG=$(go list -f {{` + `"'{{ .Dir }}'"` + `}} -m k8s.io/gengo/v2)

echo ">> Using ${CODEGEN_PKG}"

# code-generator does work with go.mod but makes assumptions about
# the project living in ` + "$GOPATH/src" + `. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> TEST TEST TEST"
echo ">> Temporary output directory ${TEMP_DIR}"

chmod +x ${CODEGEN_PKG}/kube_codegen.sh

${CODEGEN_PKG}/kube_codegen.sh all \
    ${CLIENT_PKG} \
    ${APIS_PKG} \
    {{ .ProjectConfig.Name }}:{{ .ProjectConfig.Version }} \
    --output-base "${TEMP_DIR}" --go-header-file "${GENGO_PKG}/boilerplate/boilerplate.go.txt"
# Copy everything back.
ls -la "${TEMP_DIR}"
ls -la "${TEMP_DIR}/${ROOT_PKG}"
cp -a "${TEMP_DIR}/${ROOT_PKG}/." "${SCRIPT_ROOT}/.."

`))
