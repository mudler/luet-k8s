/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PackageBuild is a specification for a PackageBuild resource
type PackageBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuildSpec   `json:"spec"`
	Status BuildStatus `json:"status"`
}

type Storage struct {
	Enabled    bool   `json:"enabled"`
	APIURL     string `json:"url"`
	SecretKey  string `json:"secretKey"`
	AccessID   string `json:"accessID"`
	Bucket     string `json:"bucket"`
	Path       string `json:"path"`
	FromSecret string `json:"fromSecret"`
}

type BuildOptions struct {
	Color   bool     `json:"color"`
	Spinner bool     `json:"spinner"`
	Values  []string `json:"values"`

	Emoji bool `json:"emoji"`

	Full bool `json:"full"`
	All  bool `json:"all"`

	Retry int `json:"retry"`

	Env []corev1.EnvVar `json:"environment"`

	Backend string `json:"backend"`

	Pull       bool     `json:"pull"`
	OnlyTarget bool     `json:"onlyTarget"`
	NoDeps     bool     `json:"noDeps"`
	Tree       []string `json:"tree"`
	Debug      bool     `json:"debug"`
	LiveOutput bool     `json:"liveOutput"`

	Push            bool                        `json:"push"`
	ImageRepository string                      `json:"imageRepository"`
	Compression     string                      `json:"compression"`
	Privileged      bool                        `json:"privileged"`
	Resources       corev1.ResourceRequirements `json:"resources"`
}

func (b BuildOptions) GetBackend() string {
	if b.Backend == "docker-sidecar" {
		return "docker"
	}
	if b.Backend == "" {
		return "img"
	}
	return b.Backend
}

type CreateRepositoryOptions struct {
	Enabled         bool     `json:"enabled"`
	Type            string   `json:"type"`
	Urls            []string `json:"urls"`
	TreeCompression string   `json:"treeCompression"`
	TreeFileName    string   `json:"treeFileName"`

	MetaCompression string `json:"metaCompression"`
	MetaFileName    string `json:"metaFileName"`

	PushImages  bool   `json:"pushImages"`
	ForcePush   bool   `json:"forcePush"`
	Description string `json:"description"`
	Name        string `json:"name"`
	OutputImage string `json:"image"`
}

type RegistryCredentials struct {
	Enabled    bool   `json:"enabled"`
	Registry   string `json:"registry"` // e.g. quay.io
	Username   string `json:"username"`
	Password   string `json:"password"`
	FromSecret string `json:"fromSecret"`
}

type Repository struct {
	Url      string `json:"url"`
	Path     string `json:"path"`
	Checkout string `json:"checkout"`
}

// BuildSpec is the spec for a PackageBuild resource
type BuildSpec struct {
	PackageName         string                  `json:"packageName"`
	Packages            []string                `json:"packages"`
	Repository          Repository              `json:"repository"`
	Storage             Storage                 `json:"storage"`
	Options             BuildOptions            `json:"options"`
	RegistryCredentials RegistryCredentials     `json:"registry"`
	NodeSelector        map[string]string       `json:"nodeSelector"`
	Annotations         map[string]string       `json:"annotations"`
	Labels              map[string]string       `json:"labels"`
	LuetRepository      CreateRepositoryOptions `json:"luetRepository"`

	PodScheduler string `json:"podScheduler"`
}

type BuildStatus struct {
	State string `json:"state"`
	Retry int    `json:"retry"`
	Node  string `json:"node"`
}
