module github.com/solo-io/solo-kit

go 1.13

require (
	github.com/Azure/go-autorest/autorest v0.9.3 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.1 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/bugsnag/bugsnag-go v1.5.0
	github.com/envoyproxy/go-control-plane v0.9.1
	github.com/envoyproxy/protoc-gen-validate v0.4.0
	github.com/fgrosse/zaptest v1.1.0
	github.com/frankban/quicktest v1.4.1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-test/deep v1.0.2 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/consul/sdk v0.3.0 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-retryablehttp v0.6.4 // indirect
	github.com/hashicorp/memberlist v0.1.5 // indirect
	github.com/hashicorp/serf v0.8.5 // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20191108163347-bdd38fca2cff
	github.com/hashicorp/vault/sdk v0.1.14-0.20191112033314-390e96e22eb2 // indirect
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/ilackarms/protoc-gen-doc v1.0.0
	github.com/mattn/go-zglob v0.0.3 // indirect
	github.com/miekg/dns v1.1.15 // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pierrec/lz4 v2.3.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pseudomuto/protoc-gen-doc v1.0.0 // indirect
	github.com/pseudomuto/protokit v0.2.0
	github.com/radovskyb/watcher v1.0.2
	github.com/rotisserie/eris v0.1.1
	github.com/solo-io/anyvendor v0.0.1
	github.com/solo-io/go-utils v0.17.0
	github.com/solo-io/protoc-gen-ext v0.0.10-0.20200904232101-c8cfa2d72872
	github.com/spf13/afero v1.3.4 // indirect
	go.opencensus.io v0.22.1
	go.uber.org/multierr v1.4.0
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/tools v0.0.0-20200811153730-74512f09e4b0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.23.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/code-generator v0.18.6
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// Breaking golang protobuf change
	//github.com/golang/protobuf => github.com/golang/protobuf v1.3.5

	// consul
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0
)
