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

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"
	scheme "github.com/mudler/luet-k8s/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PackageBuildsGetter has a method to return a PackageBuildInterface.
// A group's client should implement this interface.
type PackageBuildsGetter interface {
	PackageBuilds(namespace string) PackageBuildInterface
}

// PackageBuildInterface has methods to work with PackageBuild resources.
type PackageBuildInterface interface {
	Create(ctx context.Context, packageBuild *v1alpha1.PackageBuild, opts v1.CreateOptions) (*v1alpha1.PackageBuild, error)
	Update(ctx context.Context, packageBuild *v1alpha1.PackageBuild, opts v1.UpdateOptions) (*v1alpha1.PackageBuild, error)
	UpdateStatus(ctx context.Context, packageBuild *v1alpha1.PackageBuild, opts v1.UpdateOptions) (*v1alpha1.PackageBuild, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.PackageBuild, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.PackageBuildList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PackageBuild, err error)
	PackageBuildExpansion
}

// packageBuilds implements PackageBuildInterface
type packageBuilds struct {
	client rest.Interface
	ns     string
}

// newPackageBuilds returns a PackageBuilds
func newPackageBuilds(c *LuetV1alpha1Client, namespace string) *packageBuilds {
	return &packageBuilds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the packageBuild, and returns the corresponding packageBuild object, and an error if there is any.
func (c *packageBuilds) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PackageBuild, err error) {
	result = &v1alpha1.PackageBuild{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("packagebuilds").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PackageBuilds that match those selectors.
func (c *packageBuilds) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PackageBuildList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.PackageBuildList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("packagebuilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested packageBuilds.
func (c *packageBuilds) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("packagebuilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a packageBuild and creates it.  Returns the server's representation of the packageBuild, and an error, if there is any.
func (c *packageBuilds) Create(ctx context.Context, packageBuild *v1alpha1.PackageBuild, opts v1.CreateOptions) (result *v1alpha1.PackageBuild, err error) {
	result = &v1alpha1.PackageBuild{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("packagebuilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(packageBuild).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a packageBuild and updates it. Returns the server's representation of the packageBuild, and an error, if there is any.
func (c *packageBuilds) Update(ctx context.Context, packageBuild *v1alpha1.PackageBuild, opts v1.UpdateOptions) (result *v1alpha1.PackageBuild, err error) {
	result = &v1alpha1.PackageBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("packagebuilds").
		Name(packageBuild.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(packageBuild).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *packageBuilds) UpdateStatus(ctx context.Context, packageBuild *v1alpha1.PackageBuild, opts v1.UpdateOptions) (result *v1alpha1.PackageBuild, err error) {
	result = &v1alpha1.PackageBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("packagebuilds").
		Name(packageBuild.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(packageBuild).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the packageBuild and deletes it. Returns an error if one occurs.
func (c *packageBuilds) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("packagebuilds").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *packageBuilds) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("packagebuilds").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched packageBuild.
func (c *packageBuilds) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PackageBuild, err error) {
	result = &v1alpha1.PackageBuild{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("packagebuilds").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
