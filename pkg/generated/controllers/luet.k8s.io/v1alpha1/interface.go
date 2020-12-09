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
	v1alpha1 "github.com/mudler/luet-k8s/pkg/apis/luet.k8s.io/v1alpha1"
	clientset "github.com/mudler/luet-k8s/pkg/generated/clientset/versioned/typed/luet.k8s.io/v1alpha1"
	informers "github.com/mudler/luet-k8s/pkg/generated/informers/externalversions/luet.k8s.io/v1alpha1"
	"github.com/rancher/wrangler/pkg/generic"
)

type Interface interface {
	PackageBuild() PackageBuildController
}

func New(controllerManager *generic.ControllerManager, client clientset.LuetV1alpha1Interface,
	informers informers.Interface) Interface {
	return &version{
		controllerManager: controllerManager,
		client:            client,
		informers:         informers,
	}
}

type version struct {
	controllerManager *generic.ControllerManager
	informers         informers.Interface
	client            clientset.LuetV1alpha1Interface
}

func (c *version) PackageBuild() PackageBuildController {
	return NewPackageBuildController(v1alpha1.SchemeGroupVersion.WithKind("PackageBuild"), c.controllerManager, c.client, c.informers.PackageBuilds())
}