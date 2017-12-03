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
	kaos_v1 "github.com/arnaudmz/kaos/pkg/apis/kaos/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKaosRules implements KaosRuleInterface
type FakeKaosRules struct {
	Fake *FakeKaosV1
	ns   string
}

var kaosrulesResource = schema.GroupVersionResource{Group: "kaos.io", Version: "v1", Resource: "kaosrules"}

var kaosrulesKind = schema.GroupVersionKind{Group: "kaos.io", Version: "v1", Kind: "KaosRule"}

// Get takes name of the kaosRule, and returns the corresponding kaosRule object, and an error if there is any.
func (c *FakeKaosRules) Get(name string, options v1.GetOptions) (result *kaos_v1.KaosRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kaosrulesResource, c.ns, name), &kaos_v1.KaosRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kaos_v1.KaosRule), err
}

// List takes label and field selectors, and returns the list of KaosRules that match those selectors.
func (c *FakeKaosRules) List(opts v1.ListOptions) (result *kaos_v1.KaosRuleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kaosrulesResource, kaosrulesKind, c.ns, opts), &kaos_v1.KaosRuleList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kaos_v1.KaosRuleList{}
	for _, item := range obj.(*kaos_v1.KaosRuleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kaosRules.
func (c *FakeKaosRules) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kaosrulesResource, c.ns, opts))

}

// Create takes the representation of a kaosRule and creates it.  Returns the server's representation of the kaosRule, and an error, if there is any.
func (c *FakeKaosRules) Create(kaosRule *kaos_v1.KaosRule) (result *kaos_v1.KaosRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kaosrulesResource, c.ns, kaosRule), &kaos_v1.KaosRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kaos_v1.KaosRule), err
}

// Update takes the representation of a kaosRule and updates it. Returns the server's representation of the kaosRule, and an error, if there is any.
func (c *FakeKaosRules) Update(kaosRule *kaos_v1.KaosRule) (result *kaos_v1.KaosRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kaosrulesResource, c.ns, kaosRule), &kaos_v1.KaosRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kaos_v1.KaosRule), err
}

// Delete takes name of the kaosRule and deletes it. Returns an error if one occurs.
func (c *FakeKaosRules) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kaosrulesResource, c.ns, name), &kaos_v1.KaosRule{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKaosRules) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kaosrulesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kaos_v1.KaosRuleList{})
	return err
}

// Patch applies the patch and returns the patched kaosRule.
func (c *FakeKaosRules) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kaos_v1.KaosRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kaosrulesResource, c.ns, name, data, subresources...), &kaos_v1.KaosRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kaos_v1.KaosRule), err
}
