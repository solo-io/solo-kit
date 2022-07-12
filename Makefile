#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
PACKAGE_PATH:=github.com/solo-io/solo-kit
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)
GO_BUILD_FLAGS := GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64

VERSION ?= $(shell git describe --tags)
RELEASE := "false"

# If TAGGED_VERSION does exist, this is a release in CI
ifneq ($(TAGGED_VERSION),)
	RELEASE := "true"
	VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)
endif

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


DEPSGOBIN=$(shell pwd)/_output/.bin

.PHONY: update-deps
update-deps: mod-download
	mkdir -p $(DEPSGOBIN)
	chmod +x $(shell go list -f '{{ .Dir }}' -m k8s.io/code-generator)/generate-groups.sh
	GOBIN=$(DEPSGOBIN) go install github.com/solo-io/protoc-gen-ext
	GOBIN=$(DEPSGOBIN) go install github.com/sam-heilbron/protoc-gen-openapi
	GOBIN=$(DEPSGOBIN) go install golang.org/x/tools/cmd/goimports
	GOBIN=$(DEPSGOBIN) go install github.com/golang/protobuf/protoc-gen-go
	GOBIN=$(DEPSGOBIN) go install github.com/envoyproxy/protoc-gen-validate
	GOBIN=$(DEPSGOBIN) go install github.com/golang/mock/gomock
	GOBIN=$(DEPSGOBIN) go install github.com/golang/mock/mockgen
	GOBIN=$(DEPSGOBIN) go install github.com/onsi/ginkgo/ginkgo

	# clone solo's fork of code-generator, required for tests & kube type gen
	mkdir -p $(GOPATH)/src/k8s.io && \
		cd $(GOPATH)/src/k8s.io && \
		(git clone https://github.com/kubernetes/code-generator || echo "already found code-generator") && \
		cd $(GOPATH)/src/k8s.io/code-generator && \
		(git remote add solo https://github.com/solo-io/k8s-code-generator  || echo "already have remote solo") && \
		git fetch solo && \
		git checkout fixed-for-solo-kit-1-16-2 && \
		git pull

#----------------------------------------------------------------------------------
# Kubernetes Clientsets
#----------------------------------------------------------------------------------

$(OUTPUT_DIR):
	mkdir -p $@

.PHONY: clientset
clientset: $(OUTPUT_DIR) $(OUTPUT_DIR)/.clientset

$(OUTPUT_DIR)/.clientset: $(GENERATED_PROTO_FILES) $(SOURCES)

	$(shell go list -f '{{ .Dir }}' -m k8s.io/code-generator)/generate-groups.sh all \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd/client \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd \
		"solo.io:v1"
	touch $@

#----------------------------------------------------------------------------------
# Generated Code
#----------------------------------------------------------------------------------

.PHONY: generated-code
generated-code: $(OUTPUT_DIR)/.generated-code update-licenses

SUBDIRS:=pkg test
$(OUTPUT_DIR)/.generated-code:
	mkdir -p ${OUTPUT_DIR}
	rm -rf vendor_any
	go mod tidy
	PATH=$(DEPSGOBIN):$$PATH $(GO_BUILD_FLAGS) go generate ./...
	PATH=$(DEPSGOBIN):$$PATH gofmt -w $(SUBDIRS)
	PATH=$(DEPSGOBIN):$$PATH goimports -w $(SUBDIRS)
	touch $@

.PHONY: verify-envoy-protos
verify-envoy-protos:
	@echo Verifying validity of generated envoy files...
	PATH=$(DEPSGOBIN):$$PATH $(GO_BUILD_FLAGS) CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build pkg/api/external/verify.go


#----------------------------------------------------------------------------------
# Unit Tests
#----------------------------------------------------------------------------------

# '-skip=multicluster', '-regexScansFilePath' skips any filepath which includes multicluster, which is useful
# as this code is no longer used
.PHONY: test
test:
ifneq ($(RELEASE), "true")
	PATH=$(DEPSGOBIN):$$PATH ginkgo -r  -v -race -p -tags solokit -compilers=2 -skip multicluster -regexScansFilePath $(TEST_PKG)
endif

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

#----------------------------------------------------------------------------------
# Update third party licenses and check for GPL Licenses
#----------------------------------------------------------------------------------

update-licenses:
	# check for GPL licenses, if there are any, this will fail
	cd ci/oss_compliance; GO111MODULE=on go run oss_compliance.go osagen -c "GNU General Public License v2.0,GNU General Public License v3.0,GNU Lesser General Public License v2.1,GNU Lesser General Public License v3.0,GNU Affero General Public License v3.0"

	cd ci/oss_compliance; GO111MODULE=on go run oss_compliance.go osagen -s "Mozilla Public License 2.0,GNU General Public License v2.0,GNU General Public License v3.0,GNU Lesser General Public License v2.1,GNU Lesser General Public License v3.0,GNU Affero General Public License v3.0"> osa_provided.md
	cd ci/oss_compliance; GO111MODULE=on go run oss_compliance.go osagen -i "Mozilla Public License 2.0"> osa_included.md
