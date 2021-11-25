package main

import (
	"regexp"
	"strings"

	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func sanitizeName(s string) string {
	s = strings.ToLower(s)
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	processedString := reg.ReplaceAllString(s, "")
	return processedString
}

func newPackageBuild(packagename string, wait []string, repoBuild *v1alpha1.RepoBuild) *v1alpha1.PackageBuild {
	return &v1alpha1.PackageBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sanitizeName(packagename),
			Namespace: repoBuild.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(repoBuild, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "RepoBuild",
				}),
			},
		},
		Spec: v1alpha1.BuildSpec{
			PodScheduler:        repoBuild.Spec.PodScheduler,
			PackageName:         packagename,
			Repository:          repoBuild.Spec.GitRepository,
			Storage:             repoBuild.Spec.Storage,
			Options:             repoBuild.Spec.Options,
			RegistryCredentials: repoBuild.Spec.RegistryCredentials,
			NodeSelector:        repoBuild.Spec.NodeSelector,
			Annotations:         repoBuild.Spec.Annotations,
			Labels:              repoBuild.Spec.Labels,
			LuetRepository:      repoBuild.Spec.LuetRepository,
			WaitFor:             wait,
		},
	}
}
