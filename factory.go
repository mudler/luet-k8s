package main

import (
	"fmt"
	"strings"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func genMinioCLI(foo *v1alpha1.PackageBuild) []string {
	return []string{fmt.Sprintf(
		"mc alias set minio %s %s %s && mc cp --recursive %s minio/%s%s",
		"$STORAGE_API_URL",
		"$STORAGE_API_ID",
		"$STORAGE_API_KEY",
		"/build/*",
		foo.Spec.Storage.Bucket,
		foo.Spec.Storage.Path,
	)}
}

func genGitCommand(foo *v1alpha1.PackageBuild) []string {
	switch foo.Spec.Repository.Checkout {
	case "":
		return []string{fmt.Sprintf(
			"git clone %s /repository",
			foo.Spec.Repository.Url,
		)}

	default:
		return []string{fmt.Sprintf(
			"git clone %s /repository && cd /repository && git checkout -b build %s",
			foo.Spec.Repository.Url,
			foo.Spec.Repository.Checkout,
		)}
	}

}

func genLuetCommand(foo *v1alpha1.PackageBuild) []string {
	args := []string{"luet", "build", "--backend", "img", "--destination", "/build"}
	if foo.Spec.Options.Pull {
		args = append(args, "--pull")
	}
	if foo.Spec.Options.Push {
		args = append(args, "--push")
	}
	if len(foo.Spec.Options.ImageRepository) != 0 {
		args = append(args, "--image-repository", foo.Spec.Options.ImageRepository)
	}
	if foo.Spec.Options.NoDeps {
		args = append(args, "--nodeps")
	}
	if foo.Spec.Options.Clean {
		args = append(args, "--clean")
	}
	if foo.Spec.Options.OnlyTarget {
		args = append(args, "--only-target-package")
	}
	if len(foo.Spec.Options.Compression) != 0 {
		args = append(args, "--compression", foo.Spec.Options.Compression)
	}

	if foo.Spec.Options.Full {
		args = append(args, "--full")
	}

	if foo.Spec.Options.All {
		args = append(args, "--all")
	}

	args = append(args, fmt.Sprintf("--emoji=%t", foo.Spec.Options.Emoji))
	args = append(args, fmt.Sprintf("--color=%t", foo.Spec.Options.Color))
	args = append(args, fmt.Sprintf("--no-spinner=%t", !foo.Spec.Options.Spinner))

	if len(foo.Spec.Options.Tree) != 0 {
		for _, t := range foo.Spec.Options.Tree {
			args = append(args, "--tree", fmt.Sprintf("/repository%s", t))
		}
	} else {
		args = append(args, "--tree", fmt.Sprintf("/repository%s", foo.Spec.Repository.Path))
	}

	if foo.Spec.PackageName != "" {
		args = append(args, foo.Spec.PackageName)
	}

	for _, p := range foo.Spec.Packages {
		args = append(args, p)
	}

	if foo.Spec.RegistryCredentials.Enabled {
		args = append([]string{
			"img",
			"login",
			"-u",
			"$REGISTRY_USERNAME",
			"-p",
			"$REGISTRY_PASSWORD",
			"$REGISTRY_URI",
			"&&",
		}, args...)
	}
	return []string{strings.Join(args, " ")}
}

func genEnvVars(foo *v1alpha1.PackageBuild) []corev1.EnvVar {

	envs := []corev1.EnvVar{}

	addEnvFromSecret := func(name, secretName, secretKey string) {
		envs = append(envs, corev1.EnvVar{
			Name: name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: secretKey,
				},
			},
		})
	}

	addEnv := func(name, value string) {
		envs = append(envs, corev1.EnvVar{
			Name:  name,
			Value: value,
		})
	}

	if foo.Spec.RegistryCredentials.FromSecret != "" {
		addEnvFromSecret("REGISTRY_USERNAME", foo.Spec.RegistryCredentials.FromSecret, "registryUsername")
		addEnvFromSecret("REGISTRY_PASSWORD", foo.Spec.RegistryCredentials.FromSecret, "registryPassword")
		addEnvFromSecret("REGISTRY_URI", foo.Spec.RegistryCredentials.FromSecret, "registryUri")
	} else {
		addEnv("REGISTRY_USERNAME", foo.Spec.RegistryCredentials.Username)
		addEnv("REGISTRY_PASSWORD", foo.Spec.RegistryCredentials.Password)
		addEnv("REGISTRY_URI", foo.Spec.RegistryCredentials.Registry)
	}

	if foo.Spec.Storage.FromSecret != "" {
		addEnvFromSecret("STORAGE_API_URL", foo.Spec.Storage.FromSecret, "storageUrl")
		addEnvFromSecret("STORAGE_API_KEY", foo.Spec.Storage.FromSecret, "storageSecretKey")
		addEnvFromSecret("STORAGE_API_ID", foo.Spec.Storage.FromSecret, "storageAccessID")
	} else {
		addEnv("STORAGE_API_URL", foo.Spec.Storage.APIURL)
		addEnv("STORAGE_API_KEY", foo.Spec.Storage.SecretKey)
		addEnv("STORAGE_API_ID", foo.Spec.Storage.AccessID)
	}

	return envs
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newWorkload(foo *v1alpha1.PackageBuild) *corev1.Pod {
	secUID := int64(1000)
	privileged := false
	serviceAccount := false
	if foo.Spec.Options.Privileged {
		secUID = int64(0)
		privileged = true
	}
	pmount := corev1.UnmaskedProcMount

	podAnnotations := foo.Spec.Annotations
	if podAnnotations == nil {
		podAnnotations = make(map[string]string)
	}
	// Needed by img
	podAnnotations["container.apparmor.security.beta.kubernetes.io/spec-build"] = "unconfined"
	podAnnotations["container.seccomp.security.alpha.kubernetes.io/spec-build"] = "unconfined"

	envs := genEnvVars(foo)
	envs = append(envs, []corev1.EnvVar{
		{
			Name:  "TMPDIR",
			Value: "/buildpath",
		},
		{
			Name:  "USER",
			Value: "luet",
		},
	}...)
	pushContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env:             envs,
		Name:            "spec-push",
		Image:           "quay.io/mudler/luet-k8s-controller:latest",
		Command:         []string{"/bin/bash", "-ce"},
		Args:            genMinioCLI(foo),

		VolumeMounts: []corev1.VolumeMount{{
			Name:      "buildvolume",
			MountPath: "/build",
		}},
	}

	cloneContainer := corev1.Container{
		ImagePullPolicy: corev1.PullIfNotPresent,

		Name:    "spec-fetch",
		Image:   "quay.io/mudler/luet-k8s-controller:latest",
		Command: []string{"/bin/bash", "-cxe"},
		Args:    genGitCommand(foo),

		VolumeMounts: []corev1.VolumeMount{{
			Name:      "repository",
			MountPath: "/repository",
		}},
	}

	buildContainer := corev1.Container{
		Resources: foo.Spec.Options.Resources,
		Env:       envs,

		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  &secUID,
			ProcMount:  &pmount,
			Privileged: &privileged,
		},
		ImagePullPolicy: corev1.PullAlways,
		Name:            "spec-build",
		Image:           "quay.io/mudler/luet-k8s-controller:latest", // https://github.com/genuinetools/img/issues/289#issuecomment-626501410
		Command:         []string{"/bin/bash", "-ce"},
		Args:            genLuetCommand(foo),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "buildvolume",
				MountPath: "/build",
			},
			{
				Name:      "buildpath",
				MountPath: "/buildpath",
			},
			{
				Name:      "repository",
				MountPath: "/repository",
			},
		},
	}

	workloadPod := &corev1.Pod{
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
			Annotations: podAnnotations,
			Labels:      foo.Spec.Labels,
		},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: &serviceAccount,
			NodeSelector:                 foo.Spec.NodeSelector,
			SecurityContext:              &corev1.PodSecurityContext{RunAsUser: &secUID},
			RestartPolicy:                corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{
				{
					Name:         "buildvolume",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				},

				{
					Name:         "buildpath",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				},
				{
					Name:         "repository",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				},
			},
		},
	}
	if foo.Spec.Storage.Enabled {
		workloadPod.Spec.InitContainers = []corev1.Container{
			cloneContainer,
			buildContainer,
		}
		workloadPod.Spec.Containers = []corev1.Container{
			pushContainer,
		}
	} else {
		workloadPod.Spec.InitContainers = []corev1.Container{
			cloneContainer,
		}
		workloadPod.Spec.Containers = []corev1.Container{
			buildContainer,
		}
	}

	return workloadPod
}
