# solo-kit
A collection of code generation and libraries to for API development.

### Description:
- Define your declarative API in `.proto` files
- APIs are defined by top-level protobuf messages in `.proto` files
- Run Solo Kit's code generation and import generated files as libraries into your application. 
- These libraries provide an opinionated suite of tools for building a stateless, event-driven application.
- Currently only Go is supported, but other languages may be supported in the future.
- We are looking for community feedback and support!

### Examples
See `test/mock_resources.proto` and `test/generate.go` for an example of how to use solo-kit

## Build
- clone repo to gopath 
- gather dependencies: `dep ensure -v`
- use binary produced with `make solo-kit-gen` or import `cmd.Run(...)` into your own code gen code 

## Usage
- re-run whenever you change or add an api (.proto file)
- api objects generated from messages defined in protobuf files with magic comments prefixed with `@solo-kit`
- run `solo-kit-gen` recursively at the root of an `api` directory containing one or more `solo-kit.json` files
- generated files have the `.sk.go` suffix (generated test files do not include this suffix)

## `solo-kit-gen`

**Note**: `solo-kit-gen` will become deprecated in favor of a more robustly featured `solo-kit` CLI 
sometime in the near future.
```bash


./solo-kit-gen -h
Usage of ./solo-kit-gen:
  -docs
        generate docs as well (default true)
  -gogo
        compile normal gogo protos (default true)
  -i value
        import additional directories as proto roots (repeated flag, specify as many times as desired)
  -r string
        path to project absoluteRoot
  -s value
        skip generating for this directory (repeated flag, specify as many times as desired)

```
