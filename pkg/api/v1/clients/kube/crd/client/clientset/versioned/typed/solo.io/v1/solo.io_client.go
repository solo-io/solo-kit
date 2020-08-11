/*
Copyright The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type ResourcesV1Interface interface {
	RESTClient() rest.Interface
	ResourcesGetter
}

// ResourcesV1Client is used to interact with features provided by the resources.solo.io group.
type ResourcesV1Client struct {
	restClient rest.Interface
	def        crd.Crd
}

func (c *ResourcesV1Client) Resources(namespace string) ResourceInterface {
	return newResources(c, namespace, c.def)
}

// NewForConfig creates a new ResourcesV1Client for the given config.
func NewForConfig(c *rest.Config, def crd.Crd) (*ResourcesV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config, def); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ResourcesV1Client{restClient: client, def: def}, nil
}

// NewForConfigOrDie creates a new ResourcesV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config, def crd.Crd) *ResourcesV1Client {
	client, err := NewForConfig(c, def)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ResourcesV1Client for the given RESTClient.
func New(c rest.Interface, def crd.Crd) *ResourcesV1Client {
	return &ResourcesV1Client{restClient: c, def: def}
}

func setConfigDefaults(config *rest.Config, def crd.Crd) error {
	gv := def.GroupVersion()
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ResourcesV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
