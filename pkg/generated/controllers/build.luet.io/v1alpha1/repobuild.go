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

type RepoBuildHandler func(string, *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error)

type RepoBuildController interface {
	generic.ControllerMeta
	RepoBuildClient

	OnChange(ctx context.Context, name string, sync RepoBuildHandler)
	OnRemove(ctx context.Context, name string, sync RepoBuildHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() RepoBuildCache
}

type RepoBuildClient interface {
	Create(*v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error)
	Update(*v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error)
	UpdateStatus(*v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1alpha1.RepoBuild, error)
	List(namespace string, opts metav1.ListOptions) (*v1alpha1.RepoBuildList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.RepoBuild, err error)
}

type RepoBuildCache interface {
	Get(namespace, name string) (*v1alpha1.RepoBuild, error)
	List(namespace string, selector labels.Selector) ([]*v1alpha1.RepoBuild, error)

	AddIndexer(indexName string, indexer RepoBuildIndexer)
	GetByIndex(indexName, key string) ([]*v1alpha1.RepoBuild, error)
}

type RepoBuildIndexer func(obj *v1alpha1.RepoBuild) ([]string, error)

type repoBuildController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewRepoBuildController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) RepoBuildController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &repoBuildController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromRepoBuildHandlerToHandler(sync RepoBuildHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1alpha1.RepoBuild
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1alpha1.RepoBuild))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *repoBuildController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1alpha1.RepoBuild))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateRepoBuildDeepCopyOnChange(client RepoBuildClient, obj *v1alpha1.RepoBuild, handler func(obj *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error)) (*v1alpha1.RepoBuild, error) {
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

func (c *repoBuildController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *repoBuildController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), handler))
}

func (c *repoBuildController) OnChange(ctx context.Context, name string, sync RepoBuildHandler) {
	c.AddGenericHandler(ctx, name, FromRepoBuildHandlerToHandler(sync))
}

func (c *repoBuildController) OnRemove(ctx context.Context, name string, sync RepoBuildHandler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), FromRepoBuildHandlerToHandler(sync)))
}

func (c *repoBuildController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *repoBuildController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *repoBuildController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *repoBuildController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *repoBuildController) Cache() RepoBuildCache {
	return &repoBuildCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *repoBuildController) Create(obj *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error) {
	result := &v1alpha1.RepoBuild{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *repoBuildController) Update(obj *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error) {
	result := &v1alpha1.RepoBuild{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *repoBuildController) UpdateStatus(obj *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error) {
	result := &v1alpha1.RepoBuild{}
	return result, c.client.UpdateStatus(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *repoBuildController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *repoBuildController) Get(namespace, name string, options metav1.GetOptions) (*v1alpha1.RepoBuild, error) {
	result := &v1alpha1.RepoBuild{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *repoBuildController) List(namespace string, opts metav1.ListOptions) (*v1alpha1.RepoBuildList, error) {
	result := &v1alpha1.RepoBuildList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *repoBuildController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *repoBuildController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v1alpha1.RepoBuild, error) {
	result := &v1alpha1.RepoBuild{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type repoBuildCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *repoBuildCache) Get(namespace, name string) (*v1alpha1.RepoBuild, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1alpha1.RepoBuild), nil
}

func (c *repoBuildCache) List(namespace string, selector labels.Selector) (ret []*v1alpha1.RepoBuild, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.RepoBuild))
	})

	return ret, err
}

func (c *repoBuildCache) AddIndexer(indexName string, indexer RepoBuildIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1alpha1.RepoBuild))
		},
	}))
}

func (c *repoBuildCache) GetByIndex(indexName, key string) (result []*v1alpha1.RepoBuild, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1alpha1.RepoBuild, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1alpha1.RepoBuild))
	}
	return result, nil
}

type RepoBuildStatusHandler func(obj *v1alpha1.RepoBuild, status v1alpha1.RepoBuildStatus) (v1alpha1.RepoBuildStatus, error)

type RepoBuildGeneratingHandler func(obj *v1alpha1.RepoBuild, status v1alpha1.RepoBuildStatus) ([]runtime.Object, v1alpha1.RepoBuildStatus, error)

func RegisterRepoBuildStatusHandler(ctx context.Context, controller RepoBuildController, condition condition.Cond, name string, handler RepoBuildStatusHandler) {
	statusHandler := &repoBuildStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromRepoBuildHandlerToHandler(statusHandler.sync))
}

func RegisterRepoBuildGeneratingHandler(ctx context.Context, controller RepoBuildController, apply apply.Apply,
	condition condition.Cond, name string, handler RepoBuildGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &repoBuildGeneratingHandler{
		RepoBuildGeneratingHandler: handler,
		apply:                      apply,
		name:                       name,
		gvk:                        controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterRepoBuildStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type repoBuildStatusHandler struct {
	client    RepoBuildClient
	condition condition.Cond
	handler   RepoBuildStatusHandler
}

func (a *repoBuildStatusHandler) sync(key string, obj *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error) {
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

type repoBuildGeneratingHandler struct {
	RepoBuildGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *repoBuildGeneratingHandler) Remove(key string, obj *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1alpha1.RepoBuild{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *repoBuildGeneratingHandler) Handle(obj *v1alpha1.RepoBuild, status v1alpha1.RepoBuildStatus) (v1alpha1.RepoBuildStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.RepoBuildGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
