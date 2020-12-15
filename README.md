# (experimental) Luet Kubernetes CRD controller

Simple CRD that uses [luet](https://github.com/mudler/luet) and [img](https://github.com/genuinetools/img) to build packages on Kubernetes. It doesn't require `privileged` permissions, and builds the image as user `1000` in the workload pod.

If you need to build docker images only, have a look at [img-controller](https://github.com/mudler/img-controller/)

## Install

To install it in your k8s cluster:

```bash
$ kubectl apply -f https://raw.githubusercontent.com/mudler/luet-k8s/master/hack/kube.yaml
```

## Build packages

The controller expose a new `PackageBuild` Kubernetes resource type, which can be used to build docker images with `img` and packages with `luet`.

To build a package, for example:

```bash

$ cat <<EOF | kubectl apply -f -
apiVersion: luet.k8s.io/v1alpha1
kind: PackageBuild
metadata:
  name: test
spec:
  packageName: container/img
  repository: 
    url: "https://github.com/mocaccinoOS/mocaccino-extra"
  options:
    pull: true
    imageRepository: "quay.io/mocaccinocache/extra-amd64-cache"
EOF
```


### Full example


```yaml
apiVersion: luet.k8s.io/v1alpha1
kind: PackageBuild
metadata:
  name: test
spec:
  annotations:
    # Annotations to apply to workload pod
  labels:
    # Labels to apply to workload pod
  nodeSelector:
    # node Selector labels
  packageName: container/img
  registry:
    enabled: true
    username: "user"
    password: "pass"
    registry: "quay.io"
    fromSecret: "secret-key" # Only if using credentials from secret
  storage:
    enabled: true
    url: "minio_url"
    secretKey: "minio_secret_key"
    accessID: "minio_access_id"
    bucket: "bucket"
    path: "/bucket/path"
    fromSecret: "secret-Key" # Only if using credentials from secrets
  repository: 
    url: "https://github.com/mocaccinoOS/mocaccino-extra"
    path: "/foo/path"
    checkout: "hash_or_branch"
  options:
    pull: true
    clean: true
    onlyTarget: true
    full: true
    all: true
    privileged: true
    compression: "gzip"
    resources:
        requests:
            cpu: "100m"
            memory: "200Mi"
        limits:
            cpu: "10m"
            memory: "1Mi"
    push: true
    tree:
    - /tree/path
    noDeps: true
    color: true
    spinner: true
    imageRepository: "quay.io/mocaccinocache/extra-amd64-cache"
```

If storage and registry credentials are sourced from secrets, the secret should have the following fields and live in the same namespace of the workload:

```yaml
storageUrl: ""
storageSecretKey: ""
storageAccessID: ""
registryUri: ""
registryPassword: ""
registryUsername: ""
```

## Uninstall

First delete all the workload from the cluster, by deleting all the `packagebuild` resources.

Then run:

```bash

$ kubectl delete -f https://raw.githubusercontent.com/mudler/luet-k8s/master/hack/kube.yaml

```
