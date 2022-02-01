# Collector

This package contains code to support compiling protos

## ProtoCompiler

A ProtoCompiler is given a root and produces the file descriptor for each file in the tree. It does this by traversing the file system, starting at the root and for each file:
1. Collects the imports, using an injected Collector
2. Runs `protoc`, using an injected ProtocExecutor, and writes the output to a `tmp` file
3. Aggregates the file descriptors produced by `protoc`

## Collector

A collector is responsible for traversing a tree of files, opening and parsing each. A collector uses an `ImportsExtractor` to parse these files and aggregate a full list of imports in the tree.

## ProtocExecutor

### DefaultProtocExecutor

This is the default implementation used during codegen. It is builds the protoc command and executes it

### OpenApiProtocExecutor

This is the protoc implementation used by our `protoc` tool during `schemagen`.

## ImportsExtractor

An ImportsExtractor is responsible for reading the imports from a proto file.

### SynchronizedImportsExtractor

When extracting imports for a single file, we must walk the entire dependency tree. Since we do this for each file, it's possible that we are walking the same paths repeatedly. Therefore, we support this implementation, which only allows a single in flight request per file, and then memoizes the results.