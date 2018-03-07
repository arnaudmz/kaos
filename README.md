Kaos: Kinda Chaos Monkey for Kubernetes
=======================================

[![Docker Stars](https://img.shields.io/docker/stars/arnaudmz/kaos.svg)](https://hub.docker.com/r/arnaudmz/kaos)
[![Docker Pulls](https://img.shields.io/docker/pulls/arnaudmz/kaos.svg)](https://hub.docker.com/r/arnaudmz/kaos)
[![](https://img.shields.io/docker/automated/arnaudmz/kaos.svg)](https://hub.docker.com/r/arnaudmz/kaos)
[![ImageLayers Size](https://img.shields.io/imagelayers/image-size/arnaudmz/kaos/latest.svg)](https://hub.docker.com/r/arnaudmz/kaos)
[![ImageLayers Layers](https://img.shields.io/imagelayers/layers/arnaudmz/kaos/latest.svg)](https://hub.docker.com/r/arnaudmz/kaos)
[![Go Report Card](https://goreportcard.com/badge/github.com/arnaudmz/kaos)](https://goreportcard.com/report/github.com/arnaudmz/kaos)

Based on the CRD Custom Resources Definition examples [Kubernetes Deep Dive: Code Generation for CustomResources](https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/) and [Sample controller](https://github.com/kubernetes/sample-controller).
This code is an Operator acting as a chaos generator as Netflix [Simian Army](https://github.com/Netflix/SimianArmy).
It read chaos rules and randomly deletes matching pods. Rules are defined
using [CRD](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/):
```
apiVersion: kaos.k8s.io/v1
kind: KaosRule
metadata:
  name: my-rule
spec:
  Cron: "0 * * * * *"
  PodSelector:
    MatchLabels:
      run: nginx
```

Which will delete every minute a pod in the current namespace matching `run=nginx` selector. Cron expressions are based on https://github.com/robfig/cron implementation.

## Getting Started

First register the custom resource definition:

```
kubectl apply -f manifests/kaosrule-crd.yaml
```

Start the Operator (with its RBAC rules)

```
kubectl apply -f manifests/kaos-operator-rbac.yaml
kubectl apply -f manifests/kaos-operator-serviceaccount.yaml
kubectl apply -f manifests/kaos-operator-statefulset.yaml
```

Then add an example of the `KaosRule` kind:

```
kubectl apply -f manifests/my-rule.yaml
```

Start some matching pods to see them going down:
```
kubectl run nginx -r=8 --image=nginx:alpine
```

Build and run the example:

```
cd cmd/kaos-operator
go build
./kaos-operator -kubeconfig ~/.kube/config
```

Can also be launched as an in-cluster K8s deployment:
```
kubectl run kaos-operator --image=arnaudmz/kaos:v0.4
```

Watch the events describing kaos in action:
```
$ kubectl describe kaosrules
Name:         my-rule
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kaos.k8s.io/v1","kind":"KaosRule","metadata":{"annotations":{},"name":"my-rule","namespace":"default"},"spec":{"Cron":"0 * * * * *","Pod...
API Version:  kaos.k8s.io/v1
Kind:         KaosRule
Metadata:
  Cluster Name:
  Creation Timestamp:             2017-12-03T11:40:05Z
  Deletion Grace Period Seconds:  <nil>
  Deletion Timestamp:             <nil>
  Generation:                     0
  Initializers:                   <nil>
  Resource Version:               89920
  Self Link:                      /apis/kaos.k8s.io/v1/namespaces/default/kaosrules/my-rule
  UID:                            b4e25226-d81e-11e7-b5de-b4dde7ca9f15
Spec:
  Cron:  0 * * * * *
  Pod Selector:
    Match Labels:
      Run:  nginx
Events:
  Type     Reason             Age                From             Message
  ----     ------             ----               ----             -------
  Normal   Synced             13m                kaos-controller  Kaos Rule synced successfully and cron installed
  Warning  Pod List Empty     11m                kaos-controller  No pods matching run=nginx
  Normal   Kaos               9m                 kaos-controller  Pod nginx-5bd976694-z266d has been deleted
  Normal   Kaos               8m                 kaos-controller  Pod nginx-5bd976694-l2gmh has been deleted
```
