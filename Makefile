#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
TEMPDIR?=/tmp
PACKAGE_PATH:=github.com/solo-io/solo-kit
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)
VERSION ?= $(shell git describe --tags)

GO_BUILD_FLAGS := GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64
VENDOR=vendor

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

.PHONY: proto
proto: $(GENERATED_PROTO_FILES)

$(GENERATED_PROTO_FILES): $(PROTOS)
	cd api/v1 && \
	protoc \
	--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types:$(VENDOR)/src/ \
	-I=$(VENDOR)/src/github.com/gogo/protobuf/ \
	-I=$(VENDOR)/src/github.com/gogo/protobuf/protobuf/ \
	-I=. \
	./*.proto

.PHONY: update-deps
update-deps: vendor
	$(shell cd vendor/github.com/solo-io/protoc-gen-ext; make install)
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -u github.com/gogo/protobuf/proto
	GO111MODULE=off go get -u github.com/gogo/protobuf/jsonpb
	GO111MODULE=off go get -u github.com/gogo/protobuf/protoc-gen-gogo
	GO111MODULE=off go get -u github.com/gogo/protobuf/gogoproto
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -u github.com/golang/mock/gomock
	GO111MODULE=off go install github.com/golang/mock/mockgen

	# clone solo's fork of code-generator, required for tests & kube type gen
	mkdir -p $(VENDOR)/src/k8s.io && \
		cd $(VENDOR)/src/k8s.io && \
		(git clone https://github.com/kubernetes/code-generator || echo "already found code-generator") && \
		cd $(VENDOR)/src/k8s.io/code-generator && \
		(git remote add solo https://github.com/solo-io/k8s-code-generator  || echo "already have remote solo") && \
		git fetch solo && \
		git checkout fixed-for-solo-kit-1-16-2 && \
		git pull


.PHONY: vendor
vendor:
	go mod vendor


#----------------------------------------------------------------------------------
# Kubernetes Clientsets
#----------------------------------------------------------------------------------

$(OUTPUT_DIR):
	mkdir -p $@

.PHONY: clientset
clientset: $(OUTPUT_DIR) $(OUTPUT_DIR)/.clientset

$(OUTPUT_DIR)/.clientset: $(GENERATED_PROTO_FILES) $(SOURCES)

	$(VENDOR)/vendor/k8s.io/code-generator/generate-groups.sh all \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd/client \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd \
		"solo.io:v1"
	touch $@

#----------------------------------------------------------------------------------
# Generated Code
#----------------------------------------------------------------------------------

API_ROOT_DIR:=$(ROOTDIR)
API_IMPORTS:=\
	-I=$(API_ROOT_DIR) \
	-I=$(API_ROOT_DIR)/api/external/

GOGO_FLAG:="--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:$(TEMPDIR)"
INPUT_PROTOS=$(wildcard api/v1/*.proto)

.PHONY: solo-kit-protos
solo-kit-protos:
	protoc $(API_IMPORTS) $(GOGO_FLAG) $(INPUT_PROTOS)
	@cp -r $(TEMPDIR)/$(PACKAGE_PATH)/pkg/api/* pkg/api

.PHONY: vendor
vendor:
	go mod vendor
	chmod +x vendor/k8s.io/code-generator/generate-groups.sh

.PHONY: generated-code
generated-code: vendor solo-kit-protos $(OUTPUT_DIR)/.generated-code

SUBDIRS:=pkg test
$(OUTPUT_DIR)/.generated-code:
	mkdir -p ${OUTPUT_DIR}
	$(GO_BUILD_FLAGS) go generate ./...
	gofmt -w $(SUBDIRS)
	goimports -w $(SUBDIRS)
	touch $@

.PHONY: verify-envoy-protos
verify-envoy-protos:
	@echo Verifying validity of generated envoy files...
	$(GO_BUILD_FLAGS) CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build pkg/api/external/verify.go

#----------------------------------------------------------------------------------
# {gogo,golang}/protobuf dependencies
#----------------------------------------------------------------------------------

GOGO_PROTO_VERSION=v$(shell grep -C 1 github.com/gogo/protobuf  Gopkg.toml|grep version |cut -d'"' -f 2)
GOLANG_PROTO_VERSION=v$(shell grep -C 1 github.com/golang/protobuf  Gopkg.toml|grep version |cut -d'"' -f 2)
.PHONY: install-gen-tools
install-gogo-proto:
	mkdir -p  ${VENDOR}/src/github.com/gogo/
	mkdir -p  ${VENDOR}/src/github.com/golang/
	cd  ${VENDOR}/src/github.com/gogo/ && if [ -d protobuf ]; then cd protobuf && git fetch && git checkout $(GOGO_PROTO_VERSION); \
		else  git clone --branch $(GOGO_PROTO_VERSION) http://github.com/gogo/protobuf; fi
	cd  ${VENDOR}/src/github.com/golang/ && if [ -d protobuf ]; then cd protobuf && git fetch && git checkout $(GOLANG_PROTO_VERSION); \
		else  git clone --branch $(GOLANG_PROTO_VERSION) http://github.com/golang/protobuf; fi
	go install github.com/gogo/protobuf/protoc-gen-gogo

.PHONY: install-gen-tools
install-gen-tools: install-gogo-proto

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

