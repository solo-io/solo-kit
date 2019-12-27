# protodep

protodep is an all purpose dependency management tool originally created to manage
vendoring protobuf files. However, it can also handle any non-language specific files
available through it's multiple gathering mechanisms.

## configuration

protodep is currently only available as a library, but the plan is to turn it into a standalone tool.

To use protodep create a new protodep manager by calling `NewManager()` and supplying the working 
directory of the project. protodep is meant to work at any level of a repo/project, so therefore
the working directory must be supplied. Then the `Ensure` function can be called to vendor in 
all of the deps. The api for the `Ensure` function is reflected in the `protodep.proto` file in this 
directory.

Currently only gomod style dependencies are enabled, but git repo ones are coming soon.
### Examples

* local

```yaml
local:
    patterns:
    - **/*.proto
```
a local config can be added to any config which will vendor files directly from the current directory
into the corresponding vendor directory

* gomod

```yaml
imports:
    goMod:
      package: github.com/solo-io/solo-kit 
      patterns:
      - api/**/*.proto
```
The package is the name of the gomod package which protodep will search for the files. It will call
`go list -m all` to find the correct version, and then search the local go mod cache for it. In order to
use a package which is not explicitly required by any go projects, it can be brought in using the `tools.go`
pattern. More information on tools in go mod can be found [here](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module).


caveat: the gomod style dependency will only work if the package is specified in the list of required 
packages for a given gomod package. 

* git repo: (coming soon)


## building

Currently only involves regenerating the proto, which can be done with `go generate .`
