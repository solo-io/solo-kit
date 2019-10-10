package templates

import (
	"text/template"
)

var MultiClusterResourceClientTemplate = template.Must(template.New("multi_cluster_client").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

type {{ .Name }}MultiClusterClient interface {
	handler.ClusterHandler
	{{ .Name }}Interface
}

type {{ lower_camel .Name }}MultiClusterClient struct {
	clients       map[string]{{ .Name }}Client
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

func New{{ .Name }}MultiClusterClient(factoryGetter factory.ResourceClientFactoryGetter) {{ .Name }}MultiClusterClient {
	return New{{ .Name }}MultiClusterClientWithWatchAggregator(nil, factoryGetter)
}

func New{{ .Name }}MultiClusterClientWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) {{ .Name }}MultiClusterClient {
	return &{{ lower_camel .Name }}MultiClusterClient{
		clients:       make(map[string]{{ .Name }}Client),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *{{ lower_camel .Name }}MultiClusterClient) interfaceFor(cluster string) ({{ .Name }}Interface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, errors.Errorf("%v.%v client not found for cluster %v", "{{ .Project.ProjectConfig.Version }}", "{{ .Name }}", cluster)
}

func (c *{{ lower_camel .Name }}MultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := New{{ .Name }}Client(c.factoryGetter.ForCluster(cluster, restConfig))
	if err != nil {
		return
	}
	if err := client.Register(); err != nil {
		return
	}
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	c.clients[cluster] = client
	if c.aggregator != nil {
		c.aggregator.AddWatch(client.BaseClient())
	}
}

func (c *{{ lower_camel .Name }}MultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

{{ if .ClusterScoped }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Read(name string, opts clients.ReadOpts) (*{{ .Name }}, error) {
{{- else }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*{{ .Name }}, error) {
{{- end }}
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Read(namespace, name, opts)
}

func (c *{{ lower_camel .Name }}MultiClusterClient) Write({{ lower_camel .Name }} *{{ .Name }}, opts clients.WriteOpts) (*{{ .Name }}, error) {
	clusterInterface, err := c.interfaceFor({{ lower_camel .Name }}.GetMetadata().Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write({{ lower_camel .Name }}, opts)
}

{{ if .ClusterScoped }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Delete(name string, opts clients.DeleteOpts) error {
{{- else }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
{{- end }}
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterInterface.Delete(namespace, name, opts)
}

{{ if .ClusterScoped }}
func (c *{{ lower_camel .Name }}MultiClusterClient) List(opts clients.ListOpts) ({{ .Name }}List, error) {
{{- else }}
func (c *{{ lower_camel .Name }}MultiClusterClient) List(namespace string, opts clients.ListOpts) ({{ .Name }}List, error) {
{{- end }}
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.List(namespace, opts)
}

{{ if .ClusterScoped }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Watch(opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error) {
{{- else }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error) {
{{- end }}
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterInterface.Watch(namespace, opts)
}

`))
