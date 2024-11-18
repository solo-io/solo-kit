# https://www.gnu.org/software/make/manual/html_node/Special-Variables.html#Special-Variables
.DEFAULT_GOAL := help

#----------------------------------------------------------------------------------
# Help
#----------------------------------------------------------------------------------
# Our Makefile is quite large, and hard to reason through
# `make help` can be used to self-document targets
# To update a target to be self-documenting (and appear with the `help` command),
# place a comment after the target that is prefixed by `##`. For example:
#	custom-target: ## comment that will appear in the documentation when running `make help`
#
# **NOTE TO DEVELOPERS**
# As you encounter make targets that are frequently used, please make them self-documenting
.PHONY: help
help: FIRST_COLUMN_WIDTH=35
help: ## Output the self-documenting make targets
	@grep -hE '^[%a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-$(FIRST_COLUMN_WIDTH)s\033[0m %s\n", $$1, $$2}'

#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------
ROOTDIR := $(shell pwd)
PACKAGE_PATH:=github.com/solo-io/solo-kit
OUTPUT_DIR ?= $(ROOTDIR)/_output
DEPSGOBIN:=$(OUTPUT_DIR)/.bin
SOURCES := $(shell find . -name "*.go" | grep -v test.go)

GO_BUILD_FLAGS := GO111MODULE=on CGO_ENABLED=0

# Important to use binaries built from module.
export PATH:=$(DEPSGOBIN):$(PATH)
export GOBIN:=$(DEPSGOBIN)

#----------------------------------------------------------------------------------
# Version, Release
#----------------------------------------------------------------------------------
VERSION ?= $(shell git describe --tags --dirty | cut -c 2-)
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

.PHONY: update-all
update-all: mod-download update-deps update-code-generator

.PHONY: mod-download
mod-download:
	go mod download all

.PHONY: install-tools
install-tools: update-deps install-protoc

.PHONY: update-deps
update-deps:
	mkdir -p $(DEPSGOBIN)
	go install github.com/solo-io/protoc-gen-ext@v0.0.18
	go install github.com/solo-io/protoc-gen-openapi@v0.2.4
	go install golang.org/x/tools/cmd/goimports
	go install github.com/golang/protobuf/protoc-gen-go@v1.5.4
	go install github.com/envoyproxy/protoc-gen-validate@v1.0.4
	go install github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen

# proto compiler installation
# no explicit arm build, but x86_64 build works on arm macs
PROTOC_VERSION:=3.6.1
PROTOC_URL:=https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}
.PHONY: install-protoc
install-protoc:
	mkdir -p $(DEPSGOBIN)
	if [ $(shell ${DEPSGOBIN}/protoc --version | grep -c ${PROTOC_VERSION}) -ne 0 ]; then \
		echo expected protoc version ${PROTOC_VERSION} already installed ;\
	else \
		if [ "$(shell uname)" = "Darwin" ]; then \
			echo "downloading protoc for osx" ;\
			wget $(PROTOC_URL)-osx-x86_64.zip -O $(DEPSGOBIN)/protoc-${PROTOC_VERSION}.zip ;\
		elif [ "$(shell uname -m)" = "aarch64" ]; then \
			echo "downloading protoc for linux aarch64" ;\
			wget $(PROTOC_URL)-linux-aarch_64.zip -O $(DEPSGOBIN)/protoc-${PROTOC_VERSION}.zip ;\
		else \
			echo "downloading protoc for linux x86-64" ;\
			wget $(PROTOC_URL)-linux-x86_64.zip -O $(DEPSGOBIN)/protoc-${PROTOC_VERSION}.zip ;\
		fi ;\
		unzip $(DEPSGOBIN)/protoc-${PROTOC_VERSION}.zip -d $(DEPSGOBIN)/protoc-${PROTOC_VERSION} ;\
		mv $(DEPSGOBIN)/protoc-${PROTOC_VERSION}/bin/protoc $(DEPSGOBIN)/protoc ;\
		chmod +x $(DEPSGOBIN)/protoc ;\
		rm -rf $(DEPSGOBIN)/protoc-${PROTOC_VERSION} $(DEPSGOBIN)/protoc-${PROTOC_VERSION}.zip ;\
	fi

.PHONY: update-code-generator
update-code-generator:
	chmod +x $(shell go list -f '{{ .Dir }}' -m k8s.io/code-generator)/generate-groups.sh
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

.PHONY: clean
clean:
	rm -rf vendor_any
	find . -type d -name "doc-gen-test*" -exec rm -rf {} + # remove all doc-gen-test* directories
	find . -type d -name "_output" -exec rm -rf {} + # remove all _output directories

.PHONY: generate-all
generate-all: generated-code

.PHONY: generated-code
generated-code: $(OUTPUT_DIR)/.generated-code update-licenses

SUBDIRS:=pkg test
$(OUTPUT_DIR)/.generated-code:
	mkdir -p $(OUTPUT_DIR)
	go mod tidy
	$(GO_BUILD_FLAGS) go generate ./...
	gofmt -w $(SUBDIRS)
	$(DEPSGOBIN)/goimports -w $(SUBDIRS)
	touch $@

.PHONY: verify-envoy-protos
verify-envoy-protos:
	@echo Verifying validity of generated envoy files...
	$(GO_BUILD_FLAGS) CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build pkg/api/external/verify.go

#----------------------------------------------------------------------------------
# Tests
#----------------------------------------------------------------------------------

GINKGO_VERSION ?= $(shell echo $(shell go list -m github.com/onsi/ginkgo/v2) | cut -d' ' -f2)
GINKGO_ENV ?= GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore ACK_GINKGO_DEPRECATIONS=$(GINKGO_VERSION)
GINKGO_FLAGS ?= -v -tags=purego -compilers=4 --randomize-all --trace -progress -race
GINKGO_REPORT_FLAGS ?= --json-report=test-report.json --junit-report=junit.xml -output-dir=$(OUTPUT_DIR)
GINKGO_COVERAGE_FLAGS ?= --cover --covermode=atomic --coverprofile=coverage.cov
TEST_PKG ?= ./... # Default to run all tests

# This is a way for a user executing `make test` to be able to provide flags which we do not include by default
# For example, you may want to run tests multiple times, or with various timeouts
GINKGO_USER_FLAGS ?=

.PHONY: install-test-tools
install-test-tools:
	go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)

.PHONY: test
test: install-test-tools ## Run all tests, or only run the test package at {TEST_PKG} if it is specified
ifneq ($(RELEASE), "true")
	$(GINKGO_ENV) ginkgo \
	$(GINKGO_FLAGS) $(GINKGO_REPORT_FLAGS) $(GINKGO_USER_FLAGS) \
	$(TEST_PKG)
endif

.PHONY: test-with-coverage
test-with-coverage: GINKGO_FLAGS += $(GINKGO_COVERAGE_FLAGS)
test-with-coverage: test
	go tool cover -html $(OUTPUT_DIR)/coverage.cov

#----------------------------------------------------------------------------------
# Update third party licenses and check for GPL Licenses
#----------------------------------------------------------------------------------

update-licenses:
	# check for GPL licenses, if there are any, this will fail
	cd ci/oss_compliance; GO111MODULE=on go run oss_compliance.go osagen -c "GNU General Public License v2.0,GNU General Public License v3.0,GNU Lesser General Public License v2.1,GNU Lesser General Public License v3.0,GNU Affero General Public License v3.0"

	cd ci/oss_compliance; GO111MODULE=on go run oss_compliance.go osagen -s "Mozilla Public License 2.0,GNU General Public License v2.0,GNU General Public License v3.0,GNU Lesser General Public License v2.1,GNU Lesser General Public License v3.0,GNU Affero General Public License v3.0"> osa_provided.md
	cd ci/oss_compliance; GO111MODULE=on go run oss_compliance.go osagen -i "Mozilla Public License 2.0"> osa_included.md
