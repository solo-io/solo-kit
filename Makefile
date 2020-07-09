#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
PACKAGE_PATH:=github.com/solo-io/solo-kit
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)
VERSION ?= $(shell git describe --tags)
GO_BUILD_FLAGS := GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64

#----------------------------------------------------------------------------------
# Repo init
#----------------------------------------------------------------------------------

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

#----------------------------------------------------------------------------------
# Protobufs
#----------------------------------------------------------------------------------

PROTOS := $(shell find api/v1 -name "*.proto")
GENERATED_PROTO_FILES := $(shell find pkg/api/v1/resources/core -name "*.pb.go")

# must be a seperate target so that make waits for it to complete before moving on
.PHONY: mod-download
mod-download:
	go mod download


.PHONY: update-deps
update-deps: mod-download
	chmod +x $(shell go list -f '{{ .Dir }}' -m k8s.io/code-generator)/generate-groups.sh
	go get -v golang.org/x/tools/cmd/goimports@v0.0.0-20200423205358-59e73619c742
	go get -v github.com/gogo/protobuf/gogoproto@v1.3.1
	go get -v github.com/gogo/protobuf/protoc-gen-gogo@v1.3.1
	go get -v github.com/solo-io/protoc-gen-ext@v0.0.7
	go get -v github.com/google/wire/cmd/wire@v0.4.0
	go get -v github.com/golang/mock/mockgen@v1.4.3

#----------------------------------------------------------------------------------
# Kubernetes Clientsets
#----------------------------------------------------------------------------------

$(OUTPUT_DIR):
	mkdir -p $@

.PHONY: clientset
clientset: $(OUTPUT_DIR) $(OUTPUT_DIR)/.clientset

$(OUTPUT_DIR)/.clientset: $(GENERATED_PROTO_FILES) $(SOURCES)

	$(GOPATH)/src/k8s.io/code-generator/generate-groups.sh all \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd/client \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd \
		"solo.io:v1"
	touch $@

#----------------------------------------------------------------------------------
# Generated Code
#----------------------------------------------------------------------------------

.PHONY: generated-code
generated-code: $(OUTPUT_DIR)/.generated-code

SUBDIRS:=pkg test
$(OUTPUT_DIR)/.generated-code:
	rm -rf vendor_any
	mkdir -p ${OUTPUT_DIR}
	$(GO_BUILD_FLAGS) go generate ./...
	gofmt -w $(SUBDIRS)
	goimports -w $(SUBDIRS)
	go mod tidy
	touch $@

.PHONY: verify-envoy-protos
verify-envoy-protos:
	@echo Verifying validity of generated envoy files...
	$(GO_BUILD_FLAGS) CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build pkg/api/external/verify.go

#----------------------------------------------------------------------------------
# solo-kit-gen
#----------------------------------------------------------------------------------

solo-kit-gen:
	$(GO_BUILD_FLAGS) go build -o $@ cmd/solo-kit-gen/*.go

#----------------------------------------------------------------------------------
# solo-kit-cli
#----------------------------------------------------------------------------------

solo-kit-cli:
	$(GO_BUILD_FLAGS) go build -o $@ cmd/cli/*.go

