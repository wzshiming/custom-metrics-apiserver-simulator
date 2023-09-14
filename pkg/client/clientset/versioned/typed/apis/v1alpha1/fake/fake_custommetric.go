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
	"context"

	v1alpha1 "github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/apis/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCustomMetrics implements CustomMetricInterface
type FakeCustomMetrics struct {
	Fake *FakeApisV1alpha1
}

var custommetricsResource = v1alpha1.SchemeGroupVersion.WithResource("custommetrics")

var custommetricsKind = v1alpha1.SchemeGroupVersion.WithKind("CustomMetric")

// Get takes name of the customMetric, and returns the corresponding customMetric object, and an error if there is any.
func (c *FakeCustomMetrics) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.CustomMetric, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(custommetricsResource, name), &v1alpha1.CustomMetric{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CustomMetric), err
}

// List takes label and field selectors, and returns the list of CustomMetrics that match those selectors.
func (c *FakeCustomMetrics) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.CustomMetricList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(custommetricsResource, custommetricsKind, opts), &v1alpha1.CustomMetricList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CustomMetricList{ListMeta: obj.(*v1alpha1.CustomMetricList).ListMeta}
	for _, item := range obj.(*v1alpha1.CustomMetricList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested customMetrics.
func (c *FakeCustomMetrics) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(custommetricsResource, opts))
}

// Create takes the representation of a customMetric and creates it.  Returns the server's representation of the customMetric, and an error, if there is any.
func (c *FakeCustomMetrics) Create(ctx context.Context, customMetric *v1alpha1.CustomMetric, opts v1.CreateOptions) (result *v1alpha1.CustomMetric, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(custommetricsResource, customMetric), &v1alpha1.CustomMetric{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CustomMetric), err
}

// Update takes the representation of a customMetric and updates it. Returns the server's representation of the customMetric, and an error, if there is any.
func (c *FakeCustomMetrics) Update(ctx context.Context, customMetric *v1alpha1.CustomMetric, opts v1.UpdateOptions) (result *v1alpha1.CustomMetric, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(custommetricsResource, customMetric), &v1alpha1.CustomMetric{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CustomMetric), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCustomMetrics) UpdateStatus(ctx context.Context, customMetric *v1alpha1.CustomMetric, opts v1.UpdateOptions) (*v1alpha1.CustomMetric, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(custommetricsResource, "status", customMetric), &v1alpha1.CustomMetric{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CustomMetric), err
}

// Delete takes name of the customMetric and deletes it. Returns an error if one occurs.
func (c *FakeCustomMetrics) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(custommetricsResource, name, opts), &v1alpha1.CustomMetric{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCustomMetrics) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(custommetricsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.CustomMetricList{})
	return err
}

// Patch applies the patch and returns the patched customMetric.
func (c *FakeCustomMetrics) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.CustomMetric, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(custommetricsResource, name, pt, data, subresources...), &v1alpha1.CustomMetric{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CustomMetric), err
}