#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
PACKAGE_PATH:=github.com/solo-io/solo-kit
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)
VERSION ?= $(shell git describe --tags)

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
	cd api/v1/ && \
	protoc \
	--gogo_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types:$(GOPATH)/src/ \
	-I=$(GOPATH)/src/github.com/gogo/protobuf/ \
	-I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf/ \
	-I=. \
	./*.proto


.PHONY: update-deps
update-deps:
	GO111MODULE=off go get -v -u github.com/gogo/protobuf/protoc-gen-gogo
	GO111MODULE=off go get -v -u github.com/gogo/protobuf/gogoproto
	GO111MODULE=off go get -v -u golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -v -u github.com/golang/mock/gomock
	GO111MODULE=off go install github.com/golang/mock/mockgen

	# clone solo's fork of code-generator, required for tests & kube type gen
#	mkdir -p $(GOPATH)/src/k8s.io && \
#		cd $(GOPATH)/src/k8s.io && \
#		(git clone https://github.com/kubernetes/code-generator || echo "already found code-generator") && \
#		cd $(GOPATH)/src/k8s.io/code-generator && \
#		(git remote add solo https://github.com/solo-io/k8s-code-generator  || echo "already have remote solo") && \
#		git fetch solo && \
#		git checkout fixed-for-solo-kit && \
#		git pull


#----------------------------------------------------------------------------------
# Kubernetes Clientsets
#----------------------------------------------------------------------------------

$(OUTPUT_DIR):
	mkdir -p $@

.PHONY: clientset
clientset: $(OUTPUT_DIR) $(OUTPUT_DIR)/.clientset

$(OUTPUT_DIR)/.clientset: $(GENERATED_PROTO_FILES) $(SOURCES)
	$(ROOTDIR)/vendor/k8s.io/code-generator/generate-groups.sh all \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd/client \
		$(PACKAGE_PATH)/pkg/api/v1/clients/kube/crd \
		"solo.io:v1"
	touch $@

#----------------------------------------------------------------------------------
# Generated Code
#----------------------------------------------------------------------------------

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: generated-code
generated-code: vendor $(OUTPUT_DIR)/.generated-code

SUBDIRS:=pkg test
$(OUTPUT_DIR)/.generated-code:
	chmod +x vendor/k8s.io/code-generator/generate-groups.sh
	mkdir -p ${OUTPUT_DIR}
	go generate -x ./...
	gofmt -w $(SUBDIRS)
	goimports -w $(SUBDIRS)
	touch $@

#----------------------------------------------------------------------------------
# {gogo,golang}/protobuf dependencies
#----------------------------------------------------------------------------------

GOGO_PROTO_VERSION=v$(shell grep -C 1 github.com/gogo/protobuf  Gopkg.toml|grep version |cut -d'"' -f 2)
GOLANG_PROTO_VERSION=v$(shell grep -C 1 github.com/golang/protobuf  Gopkg.toml|grep version |cut -d'"' -f 2)
.PHONY: install-gen-tools
install-gogo-proto:
	mkdir -p  ${GOPATH}/src/github.com/gogo/ 
	mkdir -p  ${GOPATH}/src/github.com/golang/ 
	cd  ${GOPATH}/src/github.com/gogo/ && if [ -d protobuf ]; then cd protobuf && git fetch && git checkout $(GOGO_PROTO_VERSION); \
		else  git clone --branch $(GOGO_PROTO_VERSION) http://github.com/gogo/protobuf; fi
	cd  ${GOPATH}/src/github.com/golang/ && if [ -d protobuf ]; then cd protobuf && git fetch && git checkout $(GOLANG_PROTO_VERSION); \
		else  git clone --branch $(GOLANG_PROTO_VERSION) http://github.com/golang/protobuf; fi
	go install github.com/gogo/protobuf/protoc-gen-gogo

.PHONY: install-gen-tools
install-gen-tools: install-gogo-proto

#----------------------------------------------------------------------------------
# solo-kit-gen
#----------------------------------------------------------------------------------

solo-kit-gen:
	go build -o $@ cmd/solo-kit-gen/*.go

#----------------------------------------------------------------------------------
# solo-kit-cli
#----------------------------------------------------------------------------------

solo-kit-cli:
	go build -o $@ cmd/cli/*.go

