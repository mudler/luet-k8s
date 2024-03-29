/*
Copyright 2019 Wrangler Sample Controller Authors

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

// Code generated by main. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/build.luet.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRepoBuilds implements RepoBuildInterface
type FakeRepoBuilds struct {
	Fake *FakeBuildV1alpha1
	ns   string
}

var repobuildsResource = schema.GroupVersionResource{Group: "build.luet.io", Version: "v1alpha1", Resource: "repobuilds"}

var repobuildsKind = schema.GroupVersionKind{Group: "build.luet.io", Version: "v1alpha1", Kind: "RepoBuild"}

// Get takes name of the repoBuild, and returns the corresponding repoBuild object, and an error if there is any.
func (c *FakeRepoBuilds) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.RepoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(repobuildsResource, c.ns, name), &v1alpha1.RepoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RepoBuild), err
}

// List takes label and field selectors, and returns the list of RepoBuilds that match those selectors.
func (c *FakeRepoBuilds) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RepoBuildList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(repobuildsResource, repobuildsKind, c.ns, opts), &v1alpha1.RepoBuildList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.RepoBuildList{ListMeta: obj.(*v1alpha1.RepoBuildList).ListMeta}
	for _, item := range obj.(*v1alpha1.RepoBuildList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested repoBuilds.
func (c *FakeRepoBuilds) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(repobuildsResource, c.ns, opts))

}

// Create takes the representation of a repoBuild and creates it.  Returns the server's representation of the repoBuild, and an error, if there is any.
func (c *FakeRepoBuilds) Create(ctx context.Context, repoBuild *v1alpha1.RepoBuild, opts v1.CreateOptions) (result *v1alpha1.RepoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(repobuildsResource, c.ns, repoBuild), &v1alpha1.RepoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RepoBuild), err
}

// Update takes the representation of a repoBuild and updates it. Returns the server's representation of the repoBuild, and an error, if there is any.
func (c *FakeRepoBuilds) Update(ctx context.Context, repoBuild *v1alpha1.RepoBuild, opts v1.UpdateOptions) (result *v1alpha1.RepoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(repobuildsResource, c.ns, repoBuild), &v1alpha1.RepoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RepoBuild), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRepoBuilds) UpdateStatus(ctx context.Context, repoBuild *v1alpha1.RepoBuild, opts v1.UpdateOptions) (*v1alpha1.RepoBuild, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(repobuildsResource, "status", c.ns, repoBuild), &v1alpha1.RepoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RepoBuild), err
}

// Delete takes name of the repoBuild and deletes it. Returns an error if one occurs.
func (c *FakeRepoBuilds) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(repobuildsResource, c.ns, name), &v1alpha1.RepoBuild{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRepoBuilds) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(repobuildsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.RepoBuildList{})
	return err
}

// Patch applies the patch and returns the patched repoBuild.
func (c *FakeRepoBuilds) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RepoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(repobuildsResource, c.ns, name, pt, data, subresources...), &v1alpha1.RepoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.RepoBuild), err
}
