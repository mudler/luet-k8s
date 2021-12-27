package main

import (
	"fmt"
	"strings"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ImgBackend           = "img"
	DockerBackend        = "docker"
	DockerSidecarBackend = "docker-sidecar"
)

func genMinioCLI(packageBuild *v1alpha1.PackageBuild) []string {
	return []string{fmt.Sprintf(
		"mc alias set minio %s %s %s && mc cp --recursive %s minio/%s%s",
		"$STORAGE_API_URL",
		"$STORAGE_API_ID",
		"$STORAGE_API_KEY",
		"/build/*",
		packageBuild.Spec.Storage.Bucket,
		packageBuild.Spec.Storage.Path,
	)}
}

func genGitCommand(packageBuild *v1alpha1.PackageBuild) []string {
	switch packageBuild.Spec.Repository.Checkout {
	case "":
		return []string{fmt.Sprintf(
			"git clone %s /repository",
			packageBuild.Spec.Repository.Url,
		)}

	default:
		return []string{fmt.Sprintf(
			"git clone %s /repository && cd /repository && git checkout -b build %s",
			packageBuild.Spec.Repository.Url,
			packageBuild.Spec.Repository.Checkout,
		)}
	}

}

func genLuetCommand(packageBuild *v1alpha1.PackageBuild) []string {
	args := []string{"luet", "--version", "&&", "luet", "build", "--destination", "/build"}
	args = append(args, "--backend", packageBuild.Spec.Options.GetBackend())

	if packageBuild.Spec.Options.Pull {
		args = append(args, "--pull")
	}
	if packageBuild.Spec.Options.Push {
		args = append(args, "--push")
	}
	if len(packageBuild.Spec.Options.ImageRepository) != 0 {
		args = append(args, "--image-repository", packageBuild.Spec.Options.ImageRepository)
	}

	if packageBuild.Spec.Options.NoDeps {
		args = append(args, "--nodeps")
	}

	if packageBuild.Spec.Options.LiveOutput {
		args = append(args, "--live-output")
	}

	if packageBuild.Spec.Options.OnlyTarget {
		args = append(args, "--only-target-package")
	}

	if packageBuild.Spec.Options.Debug {
		args = append(args, "--debug")
	}

	if len(packageBuild.Spec.Options.Compression) != 0 {
		args = append(args, "--compression", packageBuild.Spec.Options.Compression)
	}

	if packageBuild.Spec.Options.Full {
		args = append(args, "--full")
	}

	if packageBuild.Spec.Options.All {
		args = append(args, "--all")
	}

	if packageBuild.Spec.Options.PushFinalImages {
		args = append(args, "--push-final-images")
	}

	if packageBuild.Spec.Options.PushFinalImagesWithForce {
		args = append(args, "--push-final-images-force")
	}

	if packageBuild.Spec.Options.FinalImagesRepository != "" {
		args = append(args, "--push-final-images-repository", packageBuild.Spec.Options.FinalImagesRepository)
	}

	for _, v := range packageBuild.Spec.Options.Values {
		args = append(args, fmt.Sprintf("--values=/repository%s", v))
	}

	args = append(args, fmt.Sprintf("--emoji=%t", packageBuild.Spec.Options.Emoji))
	args = append(args, fmt.Sprintf("--color=%t", packageBuild.Spec.Options.Color))
	args = append(args, fmt.Sprintf("--no-spinner=%t", !packageBuild.Spec.Options.Spinner))

	if len(packageBuild.Spec.Options.Tree) != 0 {
		for _, t := range packageBuild.Spec.Options.Tree {
			args = append(args, "--tree", fmt.Sprintf("/repository%s", t))
		}
	} else {
		args = append(args, "--tree", fmt.Sprintf("/repository%s", packageBuild.Spec.Repository.Path))
	}

	if packageBuild.Spec.PackageName != "" {
		args = append(args, packageBuild.Spec.PackageName)
	}

	for _, p := range packageBuild.Spec.Packages {
		args = append(args, p)
	}

	if packageBuild.Spec.RegistryCredentials.Enabled {
		switch packageBuild.Spec.Options.GetBackend() {
		case ImgBackend:
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
		case DockerBackend:
			args = append([]string{
				// TODO: Replace this ASAP with something that pokes docker and bails out after X attempts
				"sleep", "50", "&&",
				"docker",
				"login",
				"-u",
				"$REGISTRY_USERNAME",
				"-p",
				"$REGISTRY_PASSWORD",
				"$REGISTRY_URI",
				"&&",
			}, args...)
		}
	}
	return []string{strings.Join(args, " ")}
}

func genCreateRepoCommand(packageBuild *v1alpha1.PackageBuild) []string {
	args := []string{"luet", "--version", "&&", "luet", "create-repo", "--packages", "/build"}
	args = append(args, "--backend", packageBuild.Spec.Options.GetBackend())

	if packageBuild.Spec.LuetRepository.Type == "docker" {
		args = append(args, "--output", packageBuild.Spec.LuetRepository.OutputImage)
	} else {
		args = append(args, "--output", "/build")
	}

	if packageBuild.Spec.LuetRepository.ForcePush {
		args = append(args, "--force-push")
	}

	if packageBuild.Spec.LuetRepository.PushImages {
		args = append(args, "--push-images")
	}

	for _, t := range packageBuild.Spec.LuetRepository.Urls {
		args = append(args, "--urls", t)
	}

	if len(packageBuild.Spec.LuetRepository.Description) != 0 {
		args = append(args, "--descr", packageBuild.Spec.LuetRepository.Description)
	}

	if len(packageBuild.Spec.LuetRepository.TreeCompression) != 0 {
		args = append(args, "--tree-compression", packageBuild.Spec.LuetRepository.TreeCompression)
	}

	if len(packageBuild.Spec.LuetRepository.TreeFileName) != 0 {
		args = append(args, "--tree-filename", packageBuild.Spec.LuetRepository.TreeFileName)
	}

	if len(packageBuild.Spec.LuetRepository.MetaCompression) != 0 {
		args = append(args, "--meta-compression", packageBuild.Spec.LuetRepository.MetaCompression)
	}

	if len(packageBuild.Spec.LuetRepository.MetaFileName) != 0 {
		args = append(args, "--meta-filename", packageBuild.Spec.LuetRepository.MetaFileName)
	}

	if len(packageBuild.Spec.LuetRepository.Name) != 0 {
		args = append(args, "--name", packageBuild.Spec.LuetRepository.Name)
	}

	if len(packageBuild.Spec.LuetRepository.Type) != 0 {
		args = append(args, "--type", packageBuild.Spec.LuetRepository.Type)
	}

	if packageBuild.Spec.Options.Debug {
		args = append(args, "--debug")
	}

	args = append(args, fmt.Sprintf("--emoji=%t", packageBuild.Spec.Options.Emoji))
	args = append(args, fmt.Sprintf("--color=%t", packageBuild.Spec.Options.Color))
	args = append(args, fmt.Sprintf("--no-spinner=%t", !packageBuild.Spec.Options.Spinner))

	if len(packageBuild.Spec.Options.Tree) != 0 {
		for _, t := range packageBuild.Spec.Options.Tree {
			args = append(args, "--tree", fmt.Sprintf("/repository%s", t))
		}
	} else {
		args = append(args, "--tree", fmt.Sprintf("/repository%s", packageBuild.Spec.Repository.Path))
	}

	if packageBuild.Spec.RegistryCredentials.Enabled {
		switch packageBuild.Spec.Options.GetBackend() {
		case ImgBackend:
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
		case DockerBackend:
			args = append([]string{
				"docker",
				"login",
				"-u",
				"$REGISTRY_USERNAME",
				"-p",
				"$REGISTRY_PASSWORD",
				"$REGISTRY_URI",
				"&&",
			}, args...)
		}
	}
	return []string{strings.Join(args, " ")}
}

func genEnvVars(packageBuild *v1alpha1.PackageBuild) []corev1.EnvVar {

	envs := packageBuild.Spec.Options.Env

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

	if packageBuild.Spec.RegistryCredentials.FromSecret != "" {
		addEnvFromSecret("REGISTRY_USERNAME", packageBuild.Spec.RegistryCredentials.FromSecret, "registryUsername")
		addEnvFromSecret("REGISTRY_PASSWORD", packageBuild.Spec.RegistryCredentials.FromSecret, "registryPassword")
		addEnvFromSecret("REGISTRY_URI", packageBuild.Spec.RegistryCredentials.FromSecret, "registryUri")
	} else {
		addEnv("REGISTRY_USERNAME", packageBuild.Spec.RegistryCredentials.Username)
		addEnv("REGISTRY_PASSWORD", packageBuild.Spec.RegistryCredentials.Password)
		addEnv("REGISTRY_URI", packageBuild.Spec.RegistryCredentials.Registry)
	}

	if packageBuild.Spec.Options.Backend == DockerSidecarBackend {
		var newEnvs []corev1.EnvVar
		for _, e := range envs {
			if e.Name != "DOCKER_HOST" {
				newEnvs = append(newEnvs, e)
			}
		}
		newEnvs = append(newEnvs, corev1.EnvVar{
			Name:  "DOCKER_HOST",
			Value: "tcp://127.0.0.1:2375",
		})
		envs = newEnvs
	}

	if packageBuild.Spec.Storage.FromSecret != "" {
		addEnvFromSecret("STORAGE_API_URL", packageBuild.Spec.Storage.FromSecret, "storageUrl")
		addEnvFromSecret("STORAGE_API_KEY", packageBuild.Spec.Storage.FromSecret, "storageSecretKey")
		addEnvFromSecret("STORAGE_API_ID", packageBuild.Spec.Storage.FromSecret, "storageAccessID")
	} else {
		addEnv("STORAGE_API_URL", packageBuild.Spec.Storage.APIURL)
		addEnv("STORAGE_API_KEY", packageBuild.Spec.Storage.SecretKey)
		addEnv("STORAGE_API_ID", packageBuild.Spec.Storage.AccessID)
	}

	return envs
}

// newDeployment creates a new Deployment for a packageBuild resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the packageBuild resource that 'owns' it.
func newWorkload(packageBuild *v1alpha1.PackageBuild) *corev1.Pod {
	secUID := int64(1000)
	privileged := false
	serviceAccount := false
	if packageBuild.Spec.Options.Privileged {
		secUID = int64(0)
		privileged = true
	}
	pmount := corev1.UnmaskedProcMount

	podAnnotations := packageBuild.Spec.Annotations
	if podAnnotations == nil {
		podAnnotations = make(map[string]string)
	}

	// Needed by img
	podAnnotations["container.apparmor.security.beta.kubernetes.io/spec-build"] = "unconfined"
	podAnnotations["container.seccomp.security.alpha.kubernetes.io/spec-build"] = "unconfined"

	envs := genEnvVars(packageBuild)
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

	// pushContainer := corev1.Container{
	// 	ImagePullPolicy: corev1.PullAlways,
	// 	Env:             envs,
	// 	Name:            "spec-push",
	// 	Image:           "quay.io/mudler/luet-k8s-controller:latest",
	// 	Command:         []string{"/bin/bash", "-ce"},
	// 	Args:            genMinioCLI(packageBuild),

	// 	VolumeMounts: []corev1.VolumeMount{{
	// 		Name:      "buildvolume",
	// 		MountPath: "/build",
	// 	}},
	// }

	dockerd := []string{
		"dockerd",
		fmt.Sprintf("--mtu=%s", dockerMTU),
		"--host=tcp://0.0.0.0:2375"}

	if registryCache != "" {
		dockerd = append(dockerd, fmt.Sprintf("--registry-mirror=%s", registryCache))
	}
	if insecureRegistry != "" {
		dockerd = append(dockerd, fmt.Sprintf("--insecure-registry=%s", insecureRegistry))
	}
	dockerPrivileged := true
	dockerSidecar := corev1.Container{
		SecurityContext: &corev1.SecurityContext{Privileged: &dockerPrivileged},
		ImagePullPolicy: corev1.PullAlways,
		Env: []corev1.EnvVar{
			{
				Name:  "DOCKER_TLS_CERTDIR",
				Value: "",
			},
		},
		Name:    "docker",
		Image:   dockerImage,
		Command: dockerd,

		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "modules",
				MountPath: "/lib/modules",
				ReadOnly:  true,
			},
			{
				Name:      "cgroup",
				MountPath: "/sys/fs/cgroup",
			},
			{
				Name:      "dind-storage",
				MountPath: "/var/lib/docker",
			},
		},
	}

	cloneContainer := corev1.Container{
		ImagePullPolicy: corev1.PullAlways,

		Name:    "spec-fetch",
		Image:   "quay.io/mudler/luet-k8s-controller:latest",
		Command: []string{"/bin/bash", "-cxe"},
		Args:    genGitCommand(packageBuild),
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "repository",
			MountPath: "/repository",
		}},
	}

	buildCommand := genLuetCommand(packageBuild)
	if packageBuild.Spec.Storage.Enabled {
		buildCommand = []string{strings.Join(append(buildCommand, append([]string{"&&"}, genMinioCLI(packageBuild)...)...), " ")}
	}
	if packageBuild.Spec.LuetRepository.Enabled {
		buildCommand = []string{strings.Join(append(buildCommand, append([]string{"&&"}, genCreateRepoCommand(packageBuild)...)...), " ")}
	}
	if packageBuild.Spec.Options.Backend == DockerSidecarBackend {
		buildCommand = []string{strings.Join(append(buildCommand, []string{"&&", "pkill", "docker"}...), " ")}
	}

	buildContainer := corev1.Container{
		Resources: packageBuild.Spec.Options.Resources,
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
		Args:            buildCommand,
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
			Name:      UUID(packageBuild),
			Namespace: packageBuild.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(packageBuild, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "PackageBuild",
				}),
			},
			Annotations: podAnnotations,
			Labels:      packageBuild.Spec.Labels,
		},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: &serviceAccount,
			NodeSelector:                 packageBuild.Spec.NodeSelector,
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

	if packageBuild.Spec.PodScheduler != "" {
		workloadPod.Spec.SchedulerName = packageBuild.Spec.PodScheduler
	}

	workloadPod.Spec.InitContainers = []corev1.Container{
		cloneContainer,
	}
	workloadPod.Spec.Containers = []corev1.Container{
		buildContainer,
	}

	if packageBuild.Spec.Options.Backend == DockerSidecarBackend {
		dir := corev1.HostPathDirectory
		requiredVolumes := []corev1.Volume{
			{
				Name: "modules",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/lib/modules",
						Type: &dir,
					},
				},
			},

			{
				Name: "cgroup",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/sys/fs/cgroup",
						Type: &dir,
					},
				},
			},
			{
				Name:         "dind-storage",
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
		}

		workloadPod.Spec.Volumes = append(workloadPod.Spec.Volumes, requiredVolumes...)

		workloadPod.Spec.Containers = append(workloadPod.Spec.Containers, dockerSidecar)
		shareproc := true
		workloadPod.Spec.ShareProcessNamespace = &shareproc

	}

	return workloadPod
}
