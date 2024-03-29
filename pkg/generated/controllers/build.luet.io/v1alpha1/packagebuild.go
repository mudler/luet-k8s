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

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/build.luet.io/v1alpha1"
	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type PackageBuildHandler func(string, *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error)

type PackageBuildController interface {
	generic.ControllerMeta
	PackageBuildClient

	OnChange(ctx context.Context, name string, sync PackageBuildHandler)
	OnRemove(ctx context.Context, name string, sync PackageBuildHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() PackageBuildCache
}

type PackageBuildClient interface {
	Create(*v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error)
	Update(*v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error)
	UpdateStatus(*v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1alpha1.PackageBuild, error)
	List(namespace string, opts metav1.ListOptions) (*v1alpha1.PackageBuildList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PackageBuild, err error)
}

type PackageBuildCache interface {
	Get(namespace, name string) (*v1alpha1.PackageBuild, error)
	List(namespace string, selector labels.Selector) ([]*v1alpha1.PackageBuild, error)

	AddIndexer(indexName string, indexer PackageBuildIndexer)
	GetByIndex(indexName, key string) ([]*v1alpha1.PackageBuild, error)
}

type PackageBuildIndexer func(obj *v1alpha1.PackageBuild) ([]string, error)

type packageBuildController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewPackageBuildController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) PackageBuildController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &packageBuildController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromPackageBuildHandlerToHandler(sync PackageBuildHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1alpha1.PackageBuild
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1alpha1.PackageBuild))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *packageBuildController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1alpha1.PackageBuild))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdatePackageBuildDeepCopyOnChange(client PackageBuildClient, obj *v1alpha1.PackageBuild, handler func(obj *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error)) (*v1alpha1.PackageBuild, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *packageBuildController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *packageBuildController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), handler))
}

func (c *packageBuildController) OnChange(ctx context.Context, name string, sync PackageBuildHandler) {
	c.AddGenericHandler(ctx, name, FromPackageBuildHandlerToHandler(sync))
}

func (c *packageBuildController) OnRemove(ctx context.Context, name string, sync PackageBuildHandler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), FromPackageBuildHandlerToHandler(sync)))
}

func (c *packageBuildController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *packageBuildController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *packageBuildController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *packageBuildController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *packageBuildController) Cache() PackageBuildCache {
	return &packageBuildCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *packageBuildController) Create(obj *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error) {
	result := &v1alpha1.PackageBuild{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *packageBuildController) Update(obj *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error) {
	result := &v1alpha1.PackageBuild{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *packageBuildController) UpdateStatus(obj *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error) {
	result := &v1alpha1.PackageBuild{}
	return result, c.client.UpdateStatus(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *packageBuildController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *packageBuildController) Get(namespace, name string, options metav1.GetOptions) (*v1alpha1.PackageBuild, error) {
	result := &v1alpha1.PackageBuild{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *packageBuildController) List(namespace string, opts metav1.ListOptions) (*v1alpha1.PackageBuildList, error) {
	result := &v1alpha1.PackageBuildList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *packageBuildController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *packageBuildController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v1alpha1.PackageBuild, error) {
	result := &v1alpha1.PackageBuild{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type packageBuildCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *packageBuildCache) Get(namespace, name string) (*v1alpha1.PackageBuild, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1alpha1.PackageBuild), nil
}

func (c *packageBuildCache) List(namespace string, selector labels.Selector) (ret []*v1alpha1.PackageBuild, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PackageBuild))
	})

	return ret, err
}

func (c *packageBuildCache) AddIndexer(indexName string, indexer PackageBuildIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1alpha1.PackageBuild))
		},
	}))
}

func (c *packageBuildCache) GetByIndex(indexName, key string) (result []*v1alpha1.PackageBuild, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1alpha1.PackageBuild, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1alpha1.PackageBuild))
	}
	return result, nil
}

type PackageBuildStatusHandler func(obj *v1alpha1.PackageBuild, status v1alpha1.BuildStatus) (v1alpha1.BuildStatus, error)

type PackageBuildGeneratingHandler func(obj *v1alpha1.PackageBuild, status v1alpha1.BuildStatus) ([]runtime.Object, v1alpha1.BuildStatus, error)

func RegisterPackageBuildStatusHandler(ctx context.Context, controller PackageBuildController, condition condition.Cond, name string, handler PackageBuildStatusHandler) {
	statusHandler := &packageBuildStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromPackageBuildHandlerToHandler(statusHandler.sync))
}

func RegisterPackageBuildGeneratingHandler(ctx context.Context, controller PackageBuildController, apply apply.Apply,
	condition condition.Cond, name string, handler PackageBuildGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &packageBuildGeneratingHandler{
		PackageBuildGeneratingHandler: handler,
		apply:                         apply,
		name:                          name,
		gvk:                           controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterPackageBuildStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type packageBuildStatusHandler struct {
	client    PackageBuildClient
	condition condition.Cond
	handler   PackageBuildStatusHandler
}

func (a *packageBuildStatusHandler) sync(key string, obj *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type packageBuildGeneratingHandler struct {
	PackageBuildGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *packageBuildGeneratingHandler) Remove(key string, obj *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1alpha1.PackageBuild{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *packageBuildGeneratingHandler) Handle(obj *v1alpha1.PackageBuild, status v1alpha1.BuildStatus) (v1alpha1.BuildStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.PackageBuildGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
