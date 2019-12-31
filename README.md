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
See `test/mock_resources.proto` for an example of how to use solo-kit. These protos are generated using the
root level `generate.go` file.

## Build
- clone repo locally
- gather dependencies: `go mod download`
- import `cmd.Run(...)` into your own code gen code 

## Usage
- re-run whenever you change or add an api (.proto file)
- api objects generated from messages defined in protobuf messages which have the following requirements fulfilled:
    *     option (core.solo.io.resource).short_name = "mk";
          option (core.solo.io.resource).plural_name = "mocks";
          core.solo.io.Metadata metadata = 7 [(gogoproto.nullable) = false];
          // Optional
          core.solo.io.Status status = 6 [(gogoproto.nullable) = false, (extproto.skip_hashing) = true];
- run `solo-kit-gen` recursively at the root of an `api` directory containing one or more `solo-kit.json` files
- generated files have the `.sk.go` suffix (generated test files do not include this suffix)

## upgrading to v0.12.0 (solo-kit with go.mod)

As of go 1.11, go began introducing support for go modules, it's dependency management system.
As of solo-kit 0.12.0 we will officially support running solo-kit with go.mod outside of the GOPATH.

This change has been a lot time coming, but it also means a few changes to solo-kit.

As there is no more GOPATH, we cannot rely on the GOPATH as a method of vendoring/importing `.proto` files.
This means that we needed a new way to reliably import protos outside of the GOPATH. Therefore we created
protodep. More information on that can be found [here](https://github.com/solo-io/anyvendor).

Chief among the new changes is that the local `vendor_any` folder has become the `solo-kit` source of truth for
both `.proto` files, and `solo-kit.json` files. The `GenerateOptions` struct now takes in a protodep config.
An example of this can be found in `generate.go`. This is how solo-kit know how to vendor in the protos it 
needs to run. If this config is nil a warning will be printed, as technically this is valid, but will most likely
lead to an error.
