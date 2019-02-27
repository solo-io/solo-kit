package apiserver

//go:generate protoc -I=./ -I=${GOPATH}/src/github.com/gogo/protobuf/ -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf/ -I=${GOPATH}/src --gogo_out=plugins=grpc,import_path=apiserver,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:${GOPATH}/src --plugin=protoc-gen-solo-kit=${GOPATH}/bin/protoc-gen-solo-kit api_server.proto
