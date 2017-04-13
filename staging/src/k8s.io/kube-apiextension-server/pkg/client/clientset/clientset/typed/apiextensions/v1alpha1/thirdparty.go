/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "k8s.io/kube-apiextension-server/pkg/apis/apiextensions/v1alpha1"
	scheme "k8s.io/kube-apiextension-server/pkg/client/clientset/clientset/scheme"
)

// ThirdPartiesGetter has a method to return a ThirdPartyInterface.
// A group's client should implement this interface.
type ThirdPartiesGetter interface {
	ThirdParties() ThirdPartyInterface
}

// ThirdPartyInterface has methods to work with ThirdParty resources.
type ThirdPartyInterface interface {
	Create(*v1alpha1.ThirdParty) (*v1alpha1.ThirdParty, error)
	Update(*v1alpha1.ThirdParty) (*v1alpha1.ThirdParty, error)
	UpdateStatus(*v1alpha1.ThirdParty) (*v1alpha1.ThirdParty, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ThirdParty, error)
	List(opts v1.ListOptions) (*v1alpha1.ThirdPartyList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ThirdParty, err error)
	ThirdPartyExpansion
}

// thirdParties implements ThirdPartyInterface
type thirdParties struct {
	client rest.Interface
}

// newThirdParties returns a ThirdParties
func newThirdParties(c *ApiextensionsV1alpha1Client) *thirdParties {
	return &thirdParties{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a thirdParty and creates it.  Returns the server's representation of the thirdParty, and an error, if there is any.
func (c *thirdParties) Create(thirdParty *v1alpha1.ThirdParty) (result *v1alpha1.ThirdParty, err error) {
	result = &v1alpha1.ThirdParty{}
	err = c.client.Post().
		Resource("thirdparties").
		Body(thirdParty).
		Do().
		Into(result)
	return
}

// Update takes the representation of a thirdParty and updates it. Returns the server's representation of the thirdParty, and an error, if there is any.
func (c *thirdParties) Update(thirdParty *v1alpha1.ThirdParty) (result *v1alpha1.ThirdParty, err error) {
	result = &v1alpha1.ThirdParty{}
	err = c.client.Put().
		Resource("thirdparties").
		Name(thirdParty.Name).
		Body(thirdParty).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *thirdParties) UpdateStatus(thirdParty *v1alpha1.ThirdParty) (result *v1alpha1.ThirdParty, err error) {
	result = &v1alpha1.ThirdParty{}
	err = c.client.Put().
		Resource("thirdparties").
		Name(thirdParty.Name).
		SubResource("status").
		Body(thirdParty).
		Do().
		Into(result)
	return
}

// Delete takes name of the thirdParty and deletes it. Returns an error if one occurs.
func (c *thirdParties) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("thirdparties").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *thirdParties) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("thirdparties").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the thirdParty, and returns the corresponding thirdParty object, and an error if there is any.
func (c *thirdParties) Get(name string, options v1.GetOptions) (result *v1alpha1.ThirdParty, err error) {
	result = &v1alpha1.ThirdParty{}
	err = c.client.Get().
		Resource("thirdparties").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ThirdParties that match those selectors.
func (c *thirdParties) List(opts v1.ListOptions) (result *v1alpha1.ThirdPartyList, err error) {
	result = &v1alpha1.ThirdPartyList{}
	err = c.client.Get().
		Resource("thirdparties").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested thirdParties.
func (c *thirdParties) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("thirdparties").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched thirdParty.
func (c *thirdParties) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ThirdParty, err error) {
	result = &v1alpha1.ThirdParty{}
	err = c.client.Patch(pt).
		Resource("thirdparties").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
