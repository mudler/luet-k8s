module github.com/mudler/luet-k8s

go 1.13

//replace github.com/googleapis/gnostic => github.com/google/gnostic v0.3.1

require (
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-containerregistry v0.7.0
	github.com/mudler/luet v0.0.0-20211228210804-80bc5429bc4a
	github.com/rancher/lasso v0.0.0-20210616224652-fc3ebd901c08
	github.com/rancher/wrangler v0.8.3
	github.com/sirupsen/logrus v1.8.1
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	sigs.k8s.io/controller-runtime v0.9.0-beta.0
)
