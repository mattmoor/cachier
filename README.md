# Cachier

Cachier is an experimental Kubernetes operator that watches "duck-typed"
Kubernetes resources with the form:

```yaml
metadata:
  # Standard K8s metadata

spec:
  template:
    metadata:
      # Pod metadata
    spec:
      # Pod spec
```

Numerous built-in Kubernetes resources implement this duck-type including:
Deployment, ReplicaSet, StatefulSet, DaemonSet, and Job.


As the operator processes these resources, it creates a Knative resource of type
`caching.internal.knative.dev/v1alpha1/Image` for each of the containers in the
"pod spec".

Paired with an implementation of this resource (e.g. the [`WarmImage`
`poc-cache`](https://github.com/mattmoor/warm-image/tree/poc-cache)
implementation) the latency effect of pulling images on pod starts should be
significantly mitigated.

## Try it out

While this is not intended for public use, you can try it out via:

```bash
curl https://raw.githubusercontent.com/mattmoor/cachier/master/release.yaml \
  | kubectl apply -f -

```

Then you can see the effects with:

```bash
# Start a dummy deployment
kubectl create namespace demo
kubectl -ndemo run dummy --image=ubuntu --command -- /bin/bash -c "sleep 2592054"

# See the creating image resource:
kubectl -ndemo get image -oyaml
```

This should show something like:

```yaml
kubectl -ndemo get image -oyaml
apiVersion: v1
items:
- apiVersion: caching.internal.knative.dev/v1alpha1
  kind: Image
  metadata:
    clusterName: ""
    creationTimestamp: 2018-10-08T01:58:30Z
    generateName: dummy-00-
    generation: 1
    labels:
      controller: a792cd10-ca9d-11e8-b5eb-42010af000a3
      generation: "00001"
    name: dummy-00-nn44b
    namespace: demo
    ownerReferences:
    - apiVersion: apps/v1
      blockOwnerDeletion: true
      controller: true
      kind: Deployment
      name: dummy
      uid: a792cd10-ca9d-11e8-b5eb-42010af000a3
    resourceVersion: "26625108"
    selfLink: /apis/caching.internal.knative.dev/v1alpha1/namespaces/demo/images/dummy-00-nn44b
    uid: a7963987-ca9d-11e8-b5eb-42010af000a3
  spec:
    image: ubuntu
  status: {}
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

## Configuring the resources considered

You can customize the collection of resources to which this controller applies
by changing the following flag passed to the controller binary:

```yaml
        # Add PodSpecable types here:
        - "-resource=Deployment.v1.apps"
        - "-resource=ReplicaSet.v1.apps"
        - "-resource=StatefulSet.v1.apps"
        - "-resource=DaemonSet.v1.apps"
```

These have the form `{Kind}.{version}.{group}`, so for example a resource like:

```yaml
apiVersion: foo.mattmoor.io/v1beta2
kind: Bar
```

Would be passed as: `Bar.v1beta2.foo.mattmoor.io`


## Excluding resources from consideration

You can exclude individual resources from consideration by annotating them with:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: bar
  # This annotation excludes resources from consideration.
  annotations:
    cachier.mattmoor.io/decorate: disable
```
