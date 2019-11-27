module github.com/solo-io/solo-kit

go 1.13

require (
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/emicklei/go-restful v2.11.1+incompatible // indirect
	github.com/envoyproxy/go-control-plane v0.8.0
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/fgrosse/zaptest v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/gogo/googleapis v1.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/hashicorp/consul/api v1.1.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-retryablehttp v0.5.4 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/vault v1.1.3
	github.com/iancoleman/strcase v0.0.0-20180605031248-90d371a664d6
	github.com/ilackarms/protoc-gen-doc v1.0.0
	github.com/ilackarms/protokit v0.0.0-20181231193355-ee2393f3bbf0
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pierrec/lz4 v0.0.0-20190701081048-057d66e894a4 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pseudomuto/protoc-gen-doc v1.0.0 // indirect
	github.com/radovskyb/watcher v1.0.2
	github.com/ryanuber/go-glob v0.0.0-20160226084822-572520ed46db // indirect
	github.com/solo-io/go-utils v0.10.21
	go.opencensus.io v0.22.1
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.9.1
	golang.org/x/crypto v0.0.0-20191028145041-f83a4685e152 // indirect
	golang.org/x/net v0.0.0-20191028085509-fe3aa8a45271 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20191028164358-195ce5e7f934 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6
	google.golang.org/grpc v1.24.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191121015604-11707872ac1c
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.0.0-20191121015412-41065c7a8c2a
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4 // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// consul
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0
	// github.com/hashicorp/consul => github.com/hashicorp/consul@v1.2.1/api

	//kube 1.16
	k8s.io/api => k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191121015212-c4c8f8345c7e
)
