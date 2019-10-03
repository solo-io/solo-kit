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

package fake

import (
	v2alpha1 "github.com/solo-io/solo-kit/test/mocks/v2alpha1/kube/apis/testing.solo.io/v2alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMockResources implements MockResourceInterface
type FakeMockResources struct {
	Fake *FakeTestingV2alpha1
	ns   string
}

var mockresourcesResource = schema.GroupVersionResource{Group: "testing.solo.io", Version: "v2alpha1", Resource: "mocks"}

var mockresourcesKind = schema.GroupVersionKind{Group: "testing.solo.io", Version: "v2alpha1", Kind: "MockResource"}

// Get takes name of the mockResource, and returns the corresponding mockResource object, and an error if there is any.
func (c *FakeMockResources) Get(name string, options v1.GetOptions) (result *v2alpha1.MockResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mockresourcesResource, c.ns, name), &v2alpha1.MockResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2alpha1.MockResource), err
}

// List takes label and field selectors, and returns the list of MockResources that match those selectors.
func (c *FakeMockResources) List(opts v1.ListOptions) (result *v2alpha1.MockResourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mockresourcesResource, mockresourcesKind, c.ns, opts), &v2alpha1.MockResourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v2alpha1.MockResourceList{ListMeta: obj.(*v2alpha1.MockResourceList).ListMeta}
	for _, item := range obj.(*v2alpha1.MockResourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mockResources.
func (c *FakeMockResources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mockresourcesResource, c.ns, opts))

}

// Create takes the representation of a mockResource and creates it.  Returns the server's representation of the mockResource, and an error, if there is any.
func (c *FakeMockResources) Create(mockResource *v2alpha1.MockResource) (result *v2alpha1.MockResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mockresourcesResource, c.ns, mockResource), &v2alpha1.MockResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2alpha1.MockResource), err
}

// Update takes the representation of a mockResource and updates it. Returns the server's representation of the mockResource, and an error, if there is any.
func (c *FakeMockResources) Update(mockResource *v2alpha1.MockResource) (result *v2alpha1.MockResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mockresourcesResource, c.ns, mockResource), &v2alpha1.MockResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2alpha1.MockResource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMockResources) UpdateStatus(mockResource *v2alpha1.MockResource) (*v2alpha1.MockResource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mockresourcesResource, "status", c.ns, mockResource), &v2alpha1.MockResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2alpha1.MockResource), err
}

// Delete takes name of the mockResource and deletes it. Returns an error if one occurs.
func (c *FakeMockResources) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(mockresourcesResource, c.ns, name), &v2alpha1.MockResource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMockResources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mockresourcesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v2alpha1.MockResourceList{})
	return err
}

// Patch applies the patch and returns the patched mockResource.
func (c *FakeMockResources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2alpha1.MockResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mockresourcesResource, c.ns, name, pt, data, subresources...), &v2alpha1.MockResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2alpha1.MockResource), err
}
