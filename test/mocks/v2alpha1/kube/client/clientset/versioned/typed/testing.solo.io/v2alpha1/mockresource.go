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

package v2alpha1

import (
	"context"

	v2alpha1 "github.com/solo-io/solo-kit/test/mocks/v2alpha1/kube/apis/testing.solo.io/v2alpha1"
	scheme "github.com/solo-io/solo-kit/test/mocks/v2alpha1/kube/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// MockResourcesGetter has a method to return a MockResourceInterface.
// A group's client should implement this interface.
type MockResourcesGetter interface {
	MockResources(namespace string) MockResourceInterface
}

// MockResourceInterface has methods to work with MockResource resources.
type MockResourceInterface interface {
	Create(ctx context.Context, mockResource *v2alpha1.MockResource, opts v1.CreateOptions) (*v2alpha1.MockResource, error)
	Update(ctx context.Context, mockResource *v2alpha1.MockResource, opts v1.UpdateOptions) (*v2alpha1.MockResource, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, mockResource *v2alpha1.MockResource, opts v1.UpdateOptions) (*v2alpha1.MockResource, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v2alpha1.MockResource, error)
	List(ctx context.Context, opts v1.ListOptions) (*v2alpha1.MockResourceList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v2alpha1.MockResource, err error)
	MockResourceExpansion
}

// mockResources implements MockResourceInterface
type mockResources struct {
	*gentype.ClientWithList[*v2alpha1.MockResource, *v2alpha1.MockResourceList]
}

// newMockResources returns a MockResources
func newMockResources(c *TestingV2alpha1Client, namespace string) *mockResources {
	return &mockResources{
		gentype.NewClientWithList[*v2alpha1.MockResource, *v2alpha1.MockResourceList](
			"mocks",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v2alpha1.MockResource { return &v2alpha1.MockResource{} },
			func() *v2alpha1.MockResourceList { return &v2alpha1.MockResourceList{} }),
	}
}
