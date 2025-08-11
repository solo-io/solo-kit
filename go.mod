module github.com/solo-io/solo-kit

go 1.24.6

require (
	cuelang.org/go v0.3.2
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/bugsnag/bugsnag-go v1.5.0
	github.com/envoyproxy/go-control-plane v0.13.0
	github.com/envoyproxy/protoc-gen-validate v1.1.0
	github.com/fgrosse/zaptest v1.1.0
	github.com/getkin/kin-openapi v0.80.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-test/deep v1.0.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/vault/api v1.0.5-0.20191108163347-bdd38fca2cff
	github.com/iancoleman/strcase v0.3.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo/v2 v2.22.0
	github.com/onsi/gomega v1.36.1
	github.com/pkg/errors v0.9.1
	github.com/pseudomuto/protoc-gen-doc v1.5.0
	github.com/pseudomuto/protokit v0.2.0
	github.com/radovskyb/watcher v1.0.2
	github.com/rotisserie/eris v0.1.1
	github.com/solo-io/anyvendor v0.2.0
	github.com/solo-io/go-list-licenses v0.1.4
	github.com/solo-io/go-utils v0.28.6
	github.com/solo-io/k8s-utils v0.11.0
	github.com/solo-io/protoc-gen-ext v0.1.0
	github.com/solo-io/protoc-gen-openapi v0.3.0
	go.opencensus.io v0.24.0
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.12.0
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241209162323-e6fa225c2576
	google.golang.org/grpc v1.68.1
	google.golang.org/protobuf v1.36.5
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.33.1
	k8s.io/apiextensions-apiserver v0.33.1
	k8s.io/apimachinery v0.33.1
	k8s.io/client-go v0.33.1
	k8s.io/code-generator v0.33.1
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0
	sigs.k8s.io/yaml v1.4.0
)

require (
	cel.dev/expr v0.19.1 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/armon/go-metrics v0.3.0 // indirect
	github.com/avast/retry-go v2.2.0+incompatible // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cncf/xds/go v0.0.0-20240905190251-b4127c9b8d78 // indirect
	github.com/cockroachdb/apd/v2 v2.0.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/emicklei/proto v1.6.15 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-hclog v0.9.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/serf v0.8.2 // indirect
	github.com/hashicorp/vault/sdk v0.1.14-0.20191108161836-82f2b5571044 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/pp v2.3.0+incompatible // indirect
	github.com/lyft/protoc-gen-star v0.6.0 // indirect
	github.com/lyft/protoc-gen-star/v2 v2.0.4-0.20230330145011-496ad1ac90a4 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-zglob v0.0.3 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mpvl/unique v0.0.0-20150818121801-cbe035fff7de // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cobra v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/gengo/v2 v2.0.0-20250207200755-1244d31929d7 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250318190949-c8a335a9a2ff // indirect
	sigs.k8s.io/controller-tools v0.14.0 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2

	// pin to the jwt-go fork to fix CVE.
	// using the pseudo version of github.com/form3tech-oss/jwt-go@v3.2.3 instead of the version directly,
	// to avoid error about it being used for two different module paths
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v0.0.0-20210511163231-5b2d2b5f6c34
	github.com/docker/docker => github.com/moby/moby v24.0.7+incompatible

	// consul
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0
	github.com/iancoleman/strcase => github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
)
