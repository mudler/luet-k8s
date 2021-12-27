module github.com/mudler/luet-k8s

go 1.13

//replace github.com/googleapis/gnostic => github.com/google/gnostic v0.3.1

require (
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-containerregistry v0.6.0
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/mudler/luet v0.0.0-20211031202108-9857bea5ff59
	github.com/rancher/lasso v0.0.0-20200905045615-7fcb07d6a20b
	github.com/rancher/wrangler v0.7.2
	github.com/sirupsen/logrus v1.8.1
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)
