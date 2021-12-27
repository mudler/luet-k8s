//go:generate go run pkg/codegen/cleanup/main.go
//go:generate go run pkg/codegen/main.go

package main

import (
	"context"
	"flag"

	"github.com/mudler/luet-k8s/pkg/generated/controllers/core"
	"github.com/mudler/luet-k8s/pkg/generated/controllers/luet.k8s.io"

	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	masterURL        string
	kubeconfigFile   string
	dockerMTU        string
	registryCache    string
	dockerImage      string
	insecureRegistry string
)

func init() {
	flag.StringVar(&dockerMTU, "dockermtu", "1250", "Docker sidecar mtu.")
	flag.StringVar(&dockerImage, "dockerimage", "docker:19.03-dind", "Docker sidecar image.")
	flag.StringVar(&registryCache, "registry-mirrors", "", "Docker registry mirror")
	flag.StringVar(&insecureRegistry, "insecure-registry", "", "Docker insecure-registry")

	flag.StringVar(&kubeconfigFile, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.Parse()
}

func main() {
	// set up signals so we handle the first shutdown signal gracefully
	ctx := signals.SetupSignalHandler(context.Background())

	// This will load the kubeconfig file in a style the same as kubectl
	cfg, err := kubeconfig.GetNonInteractiveClientConfig(kubeconfigFile).ClientConfig()
	if err != nil {
		logrus.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// Raw k8s client, used to events
	kubeClient := kubernetes.NewForConfigOrDie(cfg)
	// Generated apps controller
	podsfactory := core.NewFactoryFromConfigOrDie(cfg)
	// Generated sample controller
	sample := luet.NewFactoryFromConfigOrDie(cfg)

	// The typical pattern is to build all your controller/clients then just pass to each handler
	// the bare minimum of what they need.  This will eventually help with writing tests.  So
	// don't pass in something like kubeClient, apps, or sample
	Register(ctx,
		kubeClient.CoreV1().Events(""),
		podsfactory.Core().V1().Pod(),
		sample.Luet().V1alpha1().PackageBuild(),
		sample.Luet().V1alpha1().RepoBuild())

	// Start all the controllers
	if err := start.All(ctx, 2, podsfactory, sample); err != nil {
		logrus.Fatalf("Error starting: %s", err.Error())
	}

	<-ctx.Done()
}
