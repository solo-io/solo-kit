# Schemagen
A tool used to generate validation schemas for Kubernetes CRDs

## Validation Schemas
Custom Resources in Kubernetes support [defining a structural schema](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema), which is validated on writes (ie create and update). V1 CRDs (required in Kube 1.22 and beyond) require validation schemas. These schemas provide syntax validation (ie does the configuration respect the API structure).

## Options
`crdDirectory` - Path to the directory where CRDs will be read from and written to.

`jsonSchemaTool` - Name of the tool used to generate JsonSchemas, defaults to `protoc`. `cue` is also supported, but can run excessively long for deeply nested protos.

`removeDescriptionsFromSchema` - Whether to remove descriptions from validation schemas, defaults to `false`. Descriptions are a non-functional aspect of CRD validation schemas, and therefore it is recommended to not include descriptions since they can expand the size of the CRD.

`enumAsIntOrString` - Whether to assign Enum fields the `x-kubernetes-int-or-string` property which allows the value to either be an integer, or a string, defaults to `false`. Only strings are supported when this is false. Projects which consume this library may have custom enum marshalers. If the project may marshal an enum as an int or a string, the underlying schema needs to support that as well.

`messagesWithEmptySchema` - A list of message names (ie `core.solo.io.Status`) for which we should generate a validation schema that accepts all properties. This is useful when certain messages take too long to compute schemas, or are recursive.

## Implementation
This tool executes the following steps to generate validation schemas:
1. Determine if a project is configured to generate validation schemas. If not, return immediately.
2. Read all the CRDs from the provided `crdDirectory`
3. Choose the `JsonSchemaTool`. We currently support `cue` and `protoc` though `protoc` is used in Gloo Edge
4. Generate all JsonSchemas for the project
5. For each CRD
   - Match the CRD with the JsonSchema, using the GVK.
   - Write the modified CRD back to the file, using the GVK to generate the file name

**NOTE** It would be nice to move complete CRD generation into this tool, but for now we only modify the schema.

## Generation Tools
We currently support 2 separate implementations for schema generation.

### [Cue](github.com/solo-io/cue)
`Use cuelang as an intermediate language for transpiling protobuf schemas to openapi v3 with k8s structural schema constraints.`

This is our preferred implementation. It is not used in production yet, due to some performance issues with the Gloo Edge API. However, our goal is to eventually migrate our code to rely on this implementation. This code was included in this iteration so that we could compare the generated schemas between the implementations.

[cuelang/cue#944](https://github.com/cuelang/cue/discussions/944) tracks the issue we face when using cuelang with the Gloo Edge API. Specifically, as more oneof's are added to the API, the performance degrades dramatically, to the point where it takes longer than 15 minutes to generate schemas. The authors are aware of this issue and are actively working on improving it. Until it is resolved, we cannot move forward with this implementation.

**[skv2](https://github.com/solo-io/skv2) leverages Cue to build the validation schemas for CRDs.**

### [Protoc](https://github.com/solo-io/protoc-gen-openapi)
`protoc-gen-openapi is a plugin for the Google protocol buffer compiler to generate openAPI V3 spec for any given input protobuf. It runs as a protoc-gen- binary that the protobuf compiler infers from the openapi_out flag.`

This is our current implementation that is used in production, but we hope only temporarily. As described above, the latest version of cue doesn't work with the Gloo Edge API for performance reasons.

This implementation relies on the protobuf compiler being run with an openapi_out flag. For each resource proto, we run protoc with the corresponding flags, write the schemas to a temp file, parse that file and convert it into a JsonSchema.

We based our work off an [old istio tool](https://github.com/istio/tools/tree/593a41c76c5c84a4cd51a4ab0c345630c5ed30ba/openapi/protoc-gen-openapi). It has since been removed from the repo, but we cloned it, and made necessary modifications to support specific Gloo protos.