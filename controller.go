package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	helpers "github.com/mudler/luet/cmd/helpers"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"
	luetscheme "github.com/mudler/luet-k8s/pkg/generated/clientset/versioned/scheme"
	v1 "github.com/mudler/luet-k8s/pkg/generated/controllers/core/v1"
	v1alpha1controller "github.com/mudler/luet-k8s/pkg/generated/controllers/luet.k8s.io/v1alpha1"
	"github.com/mudler/luet/pkg/api/client"
	"github.com/mudler/luet/pkg/api/core/types"
	"github.com/mudler/luet/pkg/compiler"
	compilerspec "github.com/mudler/luet/pkg/compiler/types/spec"
	installer "github.com/mudler/luet/pkg/installer"
	pkg "github.com/mudler/luet/pkg/package"
	"github.com/mudler/luet/pkg/tree"
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
	repoBuild       v1alpha1controller.RepoBuildController
	repoBuildCache  v1alpha1controller.RepoBuildCache
	recorder        record.EventRecorder
}

// NewController returns a new sample controller
func Register(
	ctx context.Context,
	events typedcorev1.EventInterface,
	pods v1.PodController,
	packageBuild v1alpha1controller.PackageBuildController,
	repoBuild v1alpha1controller.RepoBuildController) {

	controller := &Handler{
		pods:            pods,
		podsCache:       pods.Cache(),
		controller:      packageBuild,
		controllerCache: packageBuild.Cache(),
		recorder:        buildEventRecorder(events),
		repoBuild:       repoBuild,
		repoBuildCache:  repoBuild.Cache(),
	}

	// Register handlers
	pods.OnChange(ctx, "luet-handler", controller.OnPodChanged)
	packageBuild.OnChange(ctx, "luet-handler", controller.OnPackageBuildChanged)
	repoBuild.OnChange(ctx, "luet-handler", controller.OnRepoBuildChanged)
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

func (h *Handler) getRepo(url, t string) (*installer.LuetSystemRepository, error) {
	fmt.Println("Retrieving remote repository packages")
	tmpdir, err := ioutil.TempDir(os.TempDir(), "ci")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)

	d := installer.NewSystemRepository(types.LuetRepository{
		Name:   "stub",
		Type:   t,
		Cached: true,
		Urls:   []string{url},
	})
	return d.Sync(types.NewContext(), false)
}

func contains(search *client.SearchResult, packagename string) (string, bool) {

	for _, p := range search.Packages {
		if p.Category+"/"+p.Name == packagename {
			return fmt.Sprintf("%s/%s@%s", p.Category, p.Name, p.Version), true
		}
	}

	return "", false

}

func (h *Handler) OnRepoBuildChanged(key string, repoBuild *v1alpha1.RepoBuild) (*v1alpha1.RepoBuild, error) {
	// packageBuild will be nil if key is deleted from cache
	if repoBuild == nil {
		return nil, nil
	}
	logrus.Infof("Reconciling '%s' ", repoBuild.Name)

	repoType := repoBuild.Spec.Repository.Type
	if repoType == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: package name must be specified", key))
		return nil, nil
	}

	urls := repoBuild.Spec.Repository.Urls
	if len(urls) == 0 {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: repository url must be specified", key))
		return nil, nil
	}

	// We need to download the repo and get the list of packages that we want to build
	if len(repoBuild.Spec.Packages) == 0 {
		n, err := ioutil.TempDir("", "")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(n)
		// Generate package list, update ourselves with the list and enqueue
		repo, err := h.getRepo(urls[0], repoType)

		if err != nil {
			return nil, err
		}

		opts := &git.CloneOptions{
			URL:           repoBuild.Spec.GitRepository.Url,
			ReferenceName: plumbing.NewBranchReferenceName(repoBuild.Spec.GitRepository.Checkout),
		}
		// repo.BuildTree

		_, err = git.PlainClone(n, false, opts)
		if err != nil {
			return nil, err
		}

		gitRepoPackages, err := client.TreePackages(filepath.Join(n, repoBuild.Spec.GitRepository.Path))
		if err != nil {
			return nil, err
		}

		list := &client.SearchResult{}
		for _, p := range repo.BuildTree.GetDatabase().World() {
			list.Packages = append(list.Packages, client.Package{
				Name:     p.GetName(),
				Category: p.GetCategory(),
				Version:  p.GetVersion(),
			})
		}
		missingPackages := []string{}

		for _, p := range gitRepoPackages.Packages {
			if !client.Packages(list.Packages).Exist(p) {
				missingPackages = append(missingPackages, p.String())
			}
		}

		repobuildCopy := repoBuild.DeepCopy()
		repoBuild.Spec.Packages = missingPackages

		_, err = h.repoBuild.Update(repobuildCopy)
		if err != nil {
			return nil, err
		}
		// Re-enqueue ourselves with the list added
		h.repoBuild.Enqueue(repoBuild.Namespace, repoBuild.Name)

		return nil, nil
	}

	// We have the list of packages that we want to build
	opts := &git.CloneOptions{
		URL:           repoBuild.Spec.GitRepository.Url,
		ReferenceName: plumbing.NewBranchReferenceName(repoBuild.Spec.GitRepository.Checkout),
	}

	missing := &client.SearchResult{}

	n, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(n)
	// repo.BuildTree
	db := pkg.NewInMemoryDatabase(false)
	generalRecipe := tree.NewCompilerRecipe(db)

	_, err = git.PlainClone(n, false, opts)
	if err != nil {
		return nil, err
	}

	err = generalRecipe.Load(filepath.Join(n, repoBuild.Spec.GitRepository.Path))
	if err != nil {
		return nil, err
	}

	compilerSpecs := compilerspec.NewLuetCompilationspecs()
	// TODO Fill from []packages

	luetCompiler := compiler.NewLuetCompiler(nil, generalRecipe.GetDatabase())

	for _, a := range repoBuild.Spec.Packages {
		pack, err := helpers.ParsePackageStr(a)
		if err != nil {
			utilruntime.HandleError(err)
			return nil, nil
		}

		spec, err := luetCompiler.FromPackage(pack)
		if err != nil {
			utilruntime.HandleError(err)
			return nil, nil
		}
		missing.Packages = append(missing.Packages, client.Package{
			Name:     pack.GetName(),
			Category: pack.GetCategory(),
			Version:  pack.GetVersion(),
		})
		compilerSpecs.Add(spec)
	}
	bt, err := luetCompiler.BuildTree(*compilerSpecs)
	if err != nil {
		return nil, err
	}

	toCreate := []*v1alpha1.PackageBuild{}
	for _, l := range bt.AllLevels() {
		// TODO: Walk level, and link l to l-1  (waitFor)
		// and generate package build specs. Set the owner to the RepoBuild
		// Include ONLY the ones that are contained in missingPackages.

		// TODO: Do not create IF the packagebuild spec is there already

		if l == 0 {
			//	newPackageBuild() //No wait

			for _, p := range bt.AllInLevel(l) {
				full, ok := contains(missing, p)
				if ok {

					toCreate = append(toCreate, newPackageBuild(full, []string{}, repoBuild))

				}
			}
		} else {
			wait := bt.AllInLevel(l - 1) // TODO There is no PV at this stage
			waitFor := []string{}
			for _, p := range wait {
				full, ok := contains(missing, p)
				if ok {
					waitFor = append(waitFor, sanitizeName(full))
				}
			}

			for _, p := range bt.AllInLevel(l) {
				full, ok := contains(missing, p)
				if ok {

					toCreate = append(toCreate, newPackageBuild(full, waitFor, repoBuild))

				}
			}
		}

		fmt.Println(strings.Join(bt.AllInLevel(l), " "))
	}

	// Cycle over toCreate, check if it's not already there, and create!
	for _, t := range toCreate {
		_, err := h.controllerCache.Get(t.Namespace, t.Name)
		if errors.IsNotFound(err) {
			h.controller.Create(t)
		}
	}

	return nil, nil
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

		// Wait for dependant package builds
		for _, parent := range packageBuild.Spec.WaitFor {

			b, err := h.controllerCache.Get(packageBuild.Namespace, parent)
			if err != nil {
				logrus.Infof("WaitFor not found '%s' ", parent)

				return nil, err
			}

			if b.Status.State != "Succeeded" {
				logrus.Infof("parent '%s' not ready ", parent)
				return nil, err
			}
		}

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
