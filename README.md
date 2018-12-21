# solo-kit
A collection of code generation and libraries to for API development.

### Description:
- Define your declarative API in `.proto` files
- APIs are defined by top-level protobuf messages in `.proto` files

### Examples
See `test/mock_resources.proto` and `test/generate.go` for an example of how to use solo-kit

## build
- clone repo to gopath or `go get -v -u github.com/solo-io/solo-kit`
- gather dependencies: `dep ensure -v`
- use binary produced with `go build cmd/solo-kit-gen/main.go` or import `cmd.Run` into your own code gen code (must be written in Go)

## usage
- re-run whenever you change or add an api (.proto file)
- api objects generated from messages defined in protobuf files with magic comments prefixed with `@solo-kit`
- run `solo-kit-gen` recursively at the root of an `api` directory containing one or more `project.json` files
- generated files have the `.sk.go` suffix (generated test files do not include this suffix)
