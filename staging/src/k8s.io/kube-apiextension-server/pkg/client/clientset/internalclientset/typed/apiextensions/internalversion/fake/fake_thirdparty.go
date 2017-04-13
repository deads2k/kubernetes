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

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	apiextensions "k8s.io/kube-apiextension-server/pkg/apis/apiextensions"
)

// FakeThirdParties implements ThirdPartyInterface
type FakeThirdParties struct {
	Fake *FakeApiextensions
}

var thirdpartiesResource = schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "", Resource: "thirdparties"}

func (c *FakeThirdParties) Create(thirdParty *apiextensions.ThirdParty) (result *apiextensions.ThirdParty, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(thirdpartiesResource, thirdParty), &apiextensions.ThirdParty{})
	if obj == nil {
		return nil, err
	}
	return obj.(*apiextensions.ThirdParty), err
}

func (c *FakeThirdParties) Update(thirdParty *apiextensions.ThirdParty) (result *apiextensions.ThirdParty, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(thirdpartiesResource, thirdParty), &apiextensions.ThirdParty{})
	if obj == nil {
		return nil, err
	}
	return obj.(*apiextensions.ThirdParty), err
}

func (c *FakeThirdParties) UpdateStatus(thirdParty *apiextensions.ThirdParty) (*apiextensions.ThirdParty, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(thirdpartiesResource, "status", thirdParty), &apiextensions.ThirdParty{})
	if obj == nil {
		return nil, err
	}
	return obj.(*apiextensions.ThirdParty), err
}

func (c *FakeThirdParties) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(thirdpartiesResource, name), &apiextensions.ThirdParty{})
	return err
}

func (c *FakeThirdParties) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(thirdpartiesResource, listOptions)

	_, err := c.Fake.Invokes(action, &apiextensions.ThirdPartyList{})
	return err
}

func (c *FakeThirdParties) Get(name string, options v1.GetOptions) (result *apiextensions.ThirdParty, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(thirdpartiesResource, name), &apiextensions.ThirdParty{})
	if obj == nil {
		return nil, err
	}
	return obj.(*apiextensions.ThirdParty), err
}

func (c *FakeThirdParties) List(opts v1.ListOptions) (result *apiextensions.ThirdPartyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(thirdpartiesResource, opts), &apiextensions.ThirdPartyList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &apiextensions.ThirdPartyList{}
	for _, item := range obj.(*apiextensions.ThirdPartyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested thirdParties.
func (c *FakeThirdParties) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(thirdpartiesResource, opts))
}

// Patch applies the patch and returns the patched thirdParty.
func (c *FakeThirdParties) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *apiextensions.ThirdParty, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(thirdpartiesResource, name, data, subresources...), &apiextensions.ThirdParty{})
	if obj == nil {
		return nil, err
	}
	return obj.(*apiextensions.ThirdParty), err
}
