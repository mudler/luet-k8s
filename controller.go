package main

import (
	"context"
	"fmt"
	"strings"

	luetscheme "github.com/mudler/luet-k8s/pkg/generated/clientset/versioned/scheme"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"
	v1 "github.com/mudler/luet-k8s/pkg/generated/controllers/core/v1"

	v1alpha1controller "github.com/mudler/luet-k8s/pkg/generated/controllers/luet.k8s.io/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

const controllerAgentName = "luet-controller"

const (
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Luet"
)

// Handler is the controller implementation for Foo resources
type Handler struct {
	pods            v1.PodClient
	podsCache       v1.PodCache
	controller      v1alpha1controller.PackageBuildController
	controllerCache v1alpha1controller.PackageBuildCache
	recorder        record.EventRecorder
}

// NewController returns a new sample controller
func Register(
	ctx context.Context,
	events typedcorev1.EventInterface,
	pods v1.PodController,
	packageBuild v1alpha1controller.PackageBuildController) {

	controller := &Handler{
		pods:            pods,
		podsCache:       pods.Cache(),
		controller:      packageBuild,
		controllerCache: packageBuild.Cache(),
		recorder:        buildEventRecorder(events),
	}

	// Register handlers
	pods.OnChange(ctx, "luet-handler", controller.OnPodChanged)
	packageBuild.OnChange(ctx, "luet-handler", controller.OnPackageBuildChanged)
}

func buildEventRecorder(events typedcorev1.EventInterface) record.EventRecorder {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(luetscheme.AddToScheme(scheme.Scheme))
	logrus.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logrus.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: events})
	return eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
}

func cleanString(s string) string {
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func UUID(foo *v1alpha1.PackageBuild) string {
	return cleanString(fmt.Sprintf("%s-%s", foo.Spec.PackageName, foo.Spec.Repository))
}

func (h *Handler) OnPackageBuildChanged(key string, foo *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error) {
	// foo will be nil if key is deleted from cache
	if foo == nil {
		return nil, nil
	}

	packageName := foo.Spec.PackageName
	if packageName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: package name must be specified", key))
		return nil, nil
	}

	repository := foo.Spec.Repository
	if repository == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: repository name must be specified", key))
		return nil, nil
	}

	// Get the deployment with the name specified in Foo.spec
	// XXX: TODO Change, now pod name === packageName
	deployment, err := h.podsCache.Get(foo.Namespace, UUID(foo))
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = h.pods.Create(newWorkload(foo))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return nil, err
	}

	// If the Deployment is not controlled by this Foo resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(deployment, foo) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		h.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
		// Notice we don't return an error here, this is intentional because an
		// error means we should retry to reconcile.  In this situation we've done all
		// we could, which was log an error.
		return nil, nil
	}

	// If this number of the replicas on the Foo resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	// if foo.Spec.Replicas != nil && *foo.Spec.Replicas != *deployment.Spec.Replicas {
	// 	logrus.Infof("Foo %s replicas: %d, deployment replicas: %d", foo.Name, *foo.Spec.Replicas, *deployment.Spec.Replicas)
	// 	deployment, err = h.deployments.Update(newDeployment(foo))
	// }

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return nil, err
	}

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	// err = h.updateFooStatus(foo, deployment)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}

// func (h *Handler) updateFooStatus(foo *luetv1alpha1.PackageBuild, deployment *appsv1.Deployment) error {
// 	// NEVER modify objects from the store. It's a read-only, local cache.
// 	// You can use DeepCopy() to make a deep copy of original object and modify this copy
// 	// Or create a copy manually for better performance
// 	fooCopy := foo.DeepCopy()
// 	fooCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
// 	// If the CustomResourceSubresources feature gate is not enabled,
// 	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
// 	// UpdateStatus will not allow changes to the Spec of the resource,
// 	// which is ideal for ensuring nothing other than resource status has been updated.
// 	_, err := h.foos.Update(fooCopy)
// 	return err
// }

func (h *Handler) OnPodChanged(key string, pod *corev1.Pod) (*corev1.Pod, error) {
	// When an item is deleted the deployment is nil, just ignore
	if pod == nil {
		return nil, nil
	}

	if ownerRef := metav1.GetControllerOf(pod); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Foo" {
			return nil, nil
		}

		foo, err := h.podsCache.Get(pod.Namespace, ownerRef.Name)
		if err != nil {
			logrus.Infof("ignoring orphaned object '%s' of foo '%s'", pod.GetSelfLink(), ownerRef.Name)
			return nil, nil
		}

		h.controller.Enqueue(foo.Namespace, foo.Name)
		return nil, nil
	}

	return nil, nil
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newWorkload(foo *v1alpha1.PackageBuild) *corev1.Pod {

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      UUID(foo),
			Namespace: foo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "PackageBuild",
				}),
			},
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
				},
			},
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
				},
			},
		},
	}
}
