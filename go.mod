module github.com/solo-io/solo-kit

go 1.13

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20180906201452-2aa6f33b730c // indirect
	github.com/Azure/azure-sdk-for-go v37.1.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/DataDog/datadog-go v3.3.0+incompatible // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/SermoDigital/jose v0.9.1 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.60.287 // indirect
	github.com/apple/foundationdb/bindings/go v0.0.0-20191214003451-5d1974539aa9 // indirect
	github.com/aws/aws-sdk-go v1.26.2 // indirect
	github.com/bxcodec/faker v2.0.1+incompatible // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/cockroachdb/cockroach-go v0.0.0-20190925194419-606b3d062051 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20191128021309-1d7a30a10f73 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/envoyproxy/go-control-plane v0.9.1
	github.com/envoyproxy/protoc-gen-validate v0.1.0
	github.com/fgrosse/zaptest v1.1.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-ldap/ldap v3.0.3+incompatible // indirect
	github.com/gocql/gocql v0.0.0-20191126110522-1982a06ad6b9 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/hashicorp/consul v1.6.2 // indirect
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/go-gcp-common v0.6.0 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-memdb v1.0.4 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-retryablehttp v0.6.4 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20191213172644-7700d38457f3 // indirect
	github.com/hashicorp/vault v1.3.0 // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20191108163347-bdd38fca2cff
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/ilackarms/protoc-gen-doc v1.0.0
	github.com/ilackarms/protokit v0.0.0-20181231193355-ee2393f3bbf0
	github.com/jackc/pgconn v1.1.0 // indirect
	github.com/jefferai/jsonx v1.0.1 // indirect
	github.com/joyent/triton-go v1.7.0 // indirect
	github.com/keybase/go-crypto v0.0.0-20190828182435-a05457805304 // indirect
	github.com/michaelklishin/rabbit-hole v1.5.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/ncw/swift v1.0.49 // indirect
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/pierrec/cmdflag v0.0.2 // indirect
	github.com/pierrec/lz4 v2.3.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/pseudomuto/protoc-gen-doc v1.0.0 // indirect
	github.com/radovskyb/watcher v1.0.2
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da // indirect
	github.com/schollz/progressbar/v2 v2.12.1 // indirect
	github.com/solo-io/go-utils v0.11.2
	github.com/solo-io/protoc-gen-ext v0.0.1
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	github.com/ugorji/go v1.1.5-pre // indirect
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	go.etcd.io/etcd v3.3.18+incompatible // indirect
	go.opencensus.io v0.22.1
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.10.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6
	google.golang.org/grpc v1.24.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191121015604-11707872ac1c
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.0.0-20191121015412-41065c7a8c2a
	k8s.io/client-go v8.0.0+incompatible
	sigs.k8s.io/yaml v1.1.0
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// consul
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0

	//kube 1.16
	k8s.io/api => k8s.io/api v0.0.0-20191004120104-195af9ec3521
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191204090712-e0e829f17bab
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191109104512-b243870e034b
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191004123735-6bff60de4370
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)
