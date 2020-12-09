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
	Enabled   bool   `json:"enabled"`
	APIURL    string `json:"url"`
	SecretKey string `json:"secretKey"`
	AccessID  string `json:"accessID"`
	Bucket    string `json:"bucket"`
	Path      string `json:"path"`
}

type BuildOptions struct {
	Pull       bool `json:"pull"`
	Clean      bool `json:"clean"`
	OnlyTarget bool `json:"onlyTarget"`
	NoDeps     bool `json:"noDeps"`

	Push            bool   `json:"push"`
	ImageRepository string `json:"imageRepository"`
}

type RegistryCredentials struct {
	Enabled  bool   `json:"enabled"`
	Registry string `json:"registry"` // e.g. quay.io
	Username string `json:"username"`
	Password string `json:"password"`
}

type Repository struct {
	Url  string `json:"url"`
	Path string `json:"path"`
}

// BuildSpec is the spec for a PackageBuild resource
type BuildSpec struct {
	PackageName         string              `json:"packageName"`
	Repository          Repository          `json:"repository"`
	Checkout            string              `json:"checkout"`
	AsRoot              bool                `json:"asRoot"`
	Storage             Storage             `json:"storage"`
	Options             BuildOptions        `json:"options"`
	RegistryCredentials RegistryCredentials `json:"registry"`
}

type BuildStatus struct {
	State string `json:"state"`
}
