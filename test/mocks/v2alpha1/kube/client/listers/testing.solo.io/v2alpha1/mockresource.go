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

// Code generated by lister-gen. DO NOT EDIT.

package v2alpha1

import (
	v2alpha1 "github.com/solo-io/solo-kit/test/mocks/v2alpha1/kube/apis/testing.solo.io/v2alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// MockResourceLister helps list MockResources.
// All objects returned here must be treated as read-only.
type MockResourceLister interface {
	// List lists all MockResources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v2alpha1.MockResource, err error)
	// MockResources returns an object that can list and get MockResources.
	MockResources(namespace string) MockResourceNamespaceLister
	MockResourceListerExpansion
}

// mockResourceLister implements the MockResourceLister interface.
type mockResourceLister struct {
	listers.ResourceIndexer[*v2alpha1.MockResource]
}

// NewMockResourceLister returns a new MockResourceLister.
func NewMockResourceLister(indexer cache.Indexer) MockResourceLister {
	return &mockResourceLister{listers.New[*v2alpha1.MockResource](indexer, v2alpha1.Resource("mockresource"))}
}

// MockResources returns an object that can list and get MockResources.
func (s *mockResourceLister) MockResources(namespace string) MockResourceNamespaceLister {
	return mockResourceNamespaceLister{listers.NewNamespaced[*v2alpha1.MockResource](s.ResourceIndexer, namespace)}
}

// MockResourceNamespaceLister helps list and get MockResources.
// All objects returned here must be treated as read-only.
type MockResourceNamespaceLister interface {
	// List lists all MockResources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v2alpha1.MockResource, err error)
	// Get retrieves the MockResource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v2alpha1.MockResource, error)
	MockResourceNamespaceListerExpansion
}

// mockResourceNamespaceLister implements the MockResourceNamespaceLister
// interface.
type mockResourceNamespaceLister struct {
	listers.ResourceIndexer[*v2alpha1.MockResource]
}
