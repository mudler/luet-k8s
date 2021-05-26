package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"
	luetscheme "github.com/mudler/luet-k8s/pkg/generated/clientset/versioned/scheme"
	v1 "github.com/mudler/luet-k8s/pkg/generated/controllers/core/v1"
	v1alpha1controller "github.com/mudler/luet-k8s/pkg/generated/controllers/luet.k8s.io/v1alpha1"

	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

const controllerAgentName = "luet-controller"
const retryAnnotation = "luet-k8s.io/retry"

const (
	// ErrResourceExists is used as part of the Event 'reason' when a packageBuild fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Luet"
)

// Handler is the controller implementation for packageBuild resources
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
	s = strings.ToLower(s)
	return s
}

func UUID(packageBuild *v1alpha1.PackageBuild) string {
	return fmt.Sprintf("%s", packageBuild.Name)
	//	return cleanString(fmt.Sprintf("%s-%s", packageBuild.Spec.PackageName, packageBuild.Spec.Repository))
}

func PodToUUID(pod *corev1.Pod) string {
	return strings.ReplaceAll(pod.Name, "packagebuild-", "")
	//	return cleanString(fmt.Sprintf("%s-%s", packageBuild.Spec.PackageName, packageBuild.Spec.Repository))
}

func (h *Handler) OnPackageBuildChanged(key string, packageBuild *v1alpha1.PackageBuild) (*v1alpha1.PackageBuild, error) {
	// packageBuild will be nil if key is deleted from cache
	if packageBuild == nil {
		return nil, nil
	}
	logrus.Infof("Reconciling '%s' ", packageBuild.Name)

	packageName := packageBuild.Spec.PackageName
	if packageName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: package name must be specified", key))
		return nil, nil
	}

	repository := packageBuild.Spec.Repository.Url
	if repository == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: repository url must be specified", key))
		return nil, nil
	}

	// Get the deployment with the name specified in packageBuild.spec
	// XXX: TODO Change, now pod name === packageName
	logrus.Infof("Getting pod for '%s' ", packageBuild.Name)

	deployment, err := h.podsCache.Get(packageBuild.Namespace, UUID(packageBuild))
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		logrus.Infof("Pod not found for '%s' ", packageBuild.Name)

		deployment, err = h.pods.Create(newWorkload(packageBuild))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return nil, err
	}

	// If the Deployment is not controlled by this packageBuild resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(deployment, packageBuild) {
		msg := fmt.Sprintf(MessageResourceExists, packageBuild.Spec.PackageName)
		logrus.Infof(msg)

		h.recorder.Event(packageBuild, corev1.EventTypeWarning, ErrResourceExists, msg)
		// Notice we don't return an error here, this is intentional because an
		// error means we should retry to reconcile.  In this situation we've done all
		// we could, which was log an error.
		return nil, nil
	}

	// If this number of the replicas on the packageBuild resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	// if packageBuild.Spec.Replicas != nil && *packageBuild.Spec.Replicas != *deployment.Spec.Replicas {
	// 	logrus.Infof("packageBuild %s replicas: %d, deployment replicas: %d", packageBuild.Name, *packageBuild.Spec.Replicas, *deployment.Spec.Replicas)
	// 	deployment, err = h.deployments.Update(newDeployment(packageBuild))
	// }

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	// if err != nil {
	// 	return nil, err
	// }

	// Finally, we update the status block of the packageBuild resource to reflect the
	// current state of the world
	err = h.updateStatus(packageBuild, deployment)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) updateStatus(packageBuild *v1alpha1.PackageBuild, pod *corev1.Pod) error {

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	packageBuildCopy := packageBuild.DeepCopy()

	packageBuildCopy.Status.State = string(pod.Status.Phase)

	for _, c := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
		if c.State.Terminated != nil && c.State.Terminated.ExitCode != 0 {
			packageBuildCopy.Status.State = "Failed"
		}
	}

	packageBuildCopy.Status.Node = pod.Spec.NodeName
	logrus.Infof("PackageBuild '%s' has status '%s'", packageBuild.Name, packageBuildCopy.Status.State)

	if ann, exists := packageBuild.Annotations[retryAnnotation]; exists && ann != "0" {
		if nretrials, err := strconv.Atoi(ann); err == nil {
			if packageBuildCopy.Status.State == "Failed" && packageBuild.Status.Retry < nretrials {
				packageBuildCopy.Status.Retry++
				if err := h.pods.Delete(pod.Namespace, pod.Name, &metav1.DeleteOptions{}); err != nil {
					logrus.Infof("Failed deleting '%s'", pod.Name)
				}
			}
		}
	}

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the packageBuild resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := h.controller.UpdateStatus(packageBuildCopy)
	return err
}

func (h *Handler) OnPodChanged(key string, pod *corev1.Pod) (*corev1.Pod, error) {
	// When an item is deleted the deployment is nil, just ignore
	if pod == nil {
		return nil, nil
	}
	logrus.Infof("Pod '%s' has changed", pod.Name)

	if ownerRef := metav1.GetControllerOf(pod); ownerRef != nil {
		// If this object is not owned by a packageBuild, we should not do anything more
		// with it.
		if ownerRef.Kind != "PackageBuild" {
			logrus.Infof("Pod '%s' not owned by PackageBuild", pod.Name)

			return nil, nil
		}

		packageBuild, err := h.podsCache.Get(pod.Namespace, pod.Name)
		if err != nil {
			logrus.Infof("ignoring orphaned object '%s' of PackageBuild '%s'", pod.GetSelfLink(), ownerRef.Name)
			return nil, nil
		}
		logrus.Infof("Enqueueing reconcile for %s", pod.Name)

		h.controller.Enqueue(packageBuild.Namespace, packageBuild.Name)
		return nil, nil
	}

	return nil, nil
}
