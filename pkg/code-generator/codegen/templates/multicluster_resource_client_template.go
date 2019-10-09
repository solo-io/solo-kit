package templates

import (
	"text/template"
)

var MultiClusterResourceClientTemplate = template.Must(template.New("multi_cluster_client").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	"k8s.io/client-go/rest"
)

type {{ .Name }}MultiClusterClient interface {
	multicluster.ClusterHandler
	{{ .Name }}Interface
}

type {{ lower_camel .Name }}MultiClusterClient struct {
	clients      map[string]{{ .Name }}Client
	clientAccess sync.RWMutex
	aggregator   wrapper.WatchAggregator
	cacheGetter  multicluster.KubeSharedCacheGetter
	opts         multicluster.KubeResourceFactoryOpts
}

func New{{ .Name }}MultiClusterClient(cacheGetter multicluster.KubeSharedCacheGetter, opts multicluster.KubeResourceFactoryOpts) {{ .Name }}MultiClusterClient {
	return New{{ .Name }}ClientWithWatchAggregator(cacheGetter, nil, opts)
}

func New{{ .Name }}MultiClusterClientWithWatchAggregator(cacheGetter multicluster.KubeSharedCacheGetter, aggregator wrapper.WatchAggregator, opts multicluster.KubeResourceFactoryOpts) {{ .Name }}MultiClusterClient {
	return &{{ lower_camel .Name }}MultiClusterClient{
		clients:      make(map[string]{{ .Name }}Interface),
		clientAccess: sync.RWMutex{},
		cacheGetter:  cacheGetter,
		aggregator:   aggregator,
		opts:         opts,
	}
}

func (c *{{ lower_camel .Name }}MultiClusterClient) clientFor(cluster string) ({{ .Name }}Interface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, multicluster.NoClientForClusterError({{ .Name }}Crd.GroupVersionKind().String(), cluster)
}

func (c *{{ lower_camel .Name }}MultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	krc := &factory.KubeResourceClientFactory{
		Cluster:            cluster,
		Crd:                {{ .Name }}Crd,
		Cfg:                restConfig,
		SharedCache:        c.cacheGetter.GetCache(cluster),
		SkipCrdCreation:    c.opts.SkipCrdCreation,
		NamespaceWhitelist: c.opts.NamespaceWhitelist,
		ResyncPeriod:       c.opts.ResyncPeriod,
	}
	client, err := New{{ .Name }}Client(krc)
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
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.Read(namespace, name, opts)
}

func (c *{{ lower_camel .Name }}MultiClusterClient) Write({{ lower_camel .Name }} *{{ .Name }}, opts clients.WriteOpts) (*{{ .Name }}, error) {
	clusterClient, err := c.clientFor({{ lower_camel .Name }}.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterClient.Write({{ lower_camel .Name }}, opts)
}

{{ if .ClusterScoped }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Delete(name string, opts clients.DeleteOpts) error {
{{- else }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
{{- end }}
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterClient.Delete(namespace, name, opts)
}

{{ if .ClusterScoped }}
func (c *{{ lower_camel .Name }}MultiClusterClient) List(opts clients.ListOpts) ({{ .Name }}List, error) {
{{- else }}
func (c *{{ lower_camel .Name }}MultiClusterClient) List(namespace string, opts clients.ListOpts) ({{ .Name }}List, error) {
{{- end }}
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.List(namespace, opts)
}

{{ if .ClusterScoped }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Watch(opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error) {
{{- else }}
func (c *{{ lower_camel .Name }}MultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error) {
{{- end }}
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterClient.Watch(namespace, opts)
}

`))
