# External Envoy Proto Definitions
A collection of protos that were replicated from Envoy, but now maintained by Solo

## Package Naming Conventions
Since these protos are replicas of Envoy protos, we attempt to name the package in a way to reflect that. Generally the structure is `solo.io.[Envoy_PACKAGE_NAME]`. This avoids package name conflicts when we import the actual Envoy protos at the same time as these solo defined protos.

### v3 Protos
Some v3 Envoy protos need to be defined in solo-kit (for example, discovery proto messages). We will face package name conflicts if we define solo replicas in multiple projects and attempt to use them at the same time. For example, Gloo uses some solo replicas of the `Envoy.config.core.v3` package. To avoid the package name conflict, we are naming this package `solo.io.kit.[Envoy_PACKAGE]`. (Notice the `kit` in the prefix).

### v3 Discovery Protos
Discovery protos (ie https://github.com/solo-io/solo-kit/blob/e87adeba6c4dcea75ab88c884d5bef2980c6f5f9/api/external/envoy/api/v2/discovery.proto) are treated slightly differently. They are purely representational. During code-gen, we use a package replace to swap out our solo defined discovery protos, with the actual go-control-plane defined proto (https://github.com/solo-io/solo-kit/blob/e87adeba6c4dcea75ab88c884d5bef2980c6f5f9/pkg/code-generator/collector/collector.go#L252). Therefore, for discovery protos, we `match the Envoy package exactly`.
