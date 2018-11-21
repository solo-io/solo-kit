# solo-kit
A collection of code generation and libraries to for API development.

### Description:
- Define your declarative API in `.proto` files
- APIs are defined by top-level protobuf messages in `.proto` files
- Run `protoc-gen-solo-kit` plugin for `protoc` to generate an application skeleton for working with those APIs as structs in your language of choice (currently only supports Go)

### Examples
See `test/mock_resources.proto` and `test/generate.sh` for an example of how to use solo-kit

## build
- clone repo to gopath
- gather dependencies: `dep ensure -v`
- install codegen plugin `make install-plugin`

## usage
- re-run whenever you change or add an api (.proto file)
- apis are represented in protobuf files with solo-kit-specific annotations
- generate the libs provided by solo-kit: `make -B generated-code`
- generated files have the `.sk.go` suffix (generated test files do not include this suffix)
