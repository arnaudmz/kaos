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
package v1

import (
	v1 "github.com/arnaudmz/kaos/pkg/apis/kaos/v1"
	scheme "github.com/arnaudmz/kaos/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KaosRulesGetter has a method to return a KaosRuleInterface.
// A group's client should implement this interface.
type KaosRulesGetter interface {
	KaosRules(namespace string) KaosRuleInterface
}

// KaosRuleInterface has methods to work with KaosRule resources.
type KaosRuleInterface interface {
	Create(*v1.KaosRule) (*v1.KaosRule, error)
	Update(*v1.KaosRule) (*v1.KaosRule, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.KaosRule, error)
	List(opts meta_v1.ListOptions) (*v1.KaosRuleList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.KaosRule, err error)
	KaosRuleExpansion
}

// kaosRules implements KaosRuleInterface
type kaosRules struct {
	client rest.Interface
	ns     string
}

// newKaosRules returns a KaosRules
func newKaosRules(c *KaosV1Client, namespace string) *kaosRules {
	return &kaosRules{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kaosRule, and returns the corresponding kaosRule object, and an error if there is any.
func (c *kaosRules) Get(name string, options meta_v1.GetOptions) (result *v1.KaosRule, err error) {
	result = &v1.KaosRule{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kaosrules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KaosRules that match those selectors.
func (c *kaosRules) List(opts meta_v1.ListOptions) (result *v1.KaosRuleList, err error) {
	result = &v1.KaosRuleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kaosrules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kaosRules.
func (c *kaosRules) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kaosrules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a kaosRule and creates it.  Returns the server's representation of the kaosRule, and an error, if there is any.
func (c *kaosRules) Create(kaosRule *v1.KaosRule) (result *v1.KaosRule, err error) {
	result = &v1.KaosRule{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kaosrules").
		Body(kaosRule).
		Do().
		Into(result)
	return
}

// Update takes the representation of a kaosRule and updates it. Returns the server's representation of the kaosRule, and an error, if there is any.
func (c *kaosRules) Update(kaosRule *v1.KaosRule) (result *v1.KaosRule, err error) {
	result = &v1.KaosRule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kaosrules").
		Name(kaosRule.Name).
		Body(kaosRule).
		Do().
		Into(result)
	return
}

// Delete takes name of the kaosRule and deletes it. Returns an error if one occurs.
func (c *kaosRules) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kaosrules").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kaosRules) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kaosrules").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched kaosRule.
func (c *kaosRules) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.KaosRule, err error) {
	result = &v1.KaosRule{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kaosrules").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
