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

.PHONY: update-deps
update-deps:
	mkdir -p $(DEPSGOBIN)
	go install github.com/solo-io/protoc-gen-ext
	go install github.com/solo-io/protoc-gen-openapi@v0.1.0
	go install golang.org/x/tools/cmd/goimports
	go install github.com/golang/protobuf/protoc-gen-go
	go install github.com/envoyproxy/protoc-gen-validate
	go install github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen

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
	rm -rf $(OUTPUT_DIR)

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
	goimports -w $(SUBDIRS)
	touch $@

.PHONY: verify-envoy-protos
verify-envoy-protos:
	@echo Verifying validity of generated envoy files...
	$(GO_BUILD_FLAGS) CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build pkg/api/external/verify.go

#----------------------------------------------------------------------------------
# Tests
#----------------------------------------------------------------------------------

GINKGO_VERSION ?= 2.5.0 # match our go.mod
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
	go install github.com/onsi/ginkgo/v2/ginkgo@v$(GINKGO_VERSION)

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
