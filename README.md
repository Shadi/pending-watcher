# pending-watcher
Small package to find and log pods that has been pending in a kubernetes cluster for more than 2 minutes

This should help in finding pods that stay pending on some nodes because of resources constraint or any other reasons,
I originally created this to find daemonset pods that don't get deployed correctly to all nodes when its updated, so
it will log which pod is stuck in pending, and the node it is trying to get deployed on, but this will work on all other pending
cases but it won't have the node name since it specifically looking for the field `NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution`
which is the important one for the original daemonset usecase.

### Install
If you have go installed and `GOPATH` in you path you can install it using go install:
```
go install -v github.com/shadi/pending-watcher@latest
```

You can also download the binary from github:
```
wget https://github.com/Shadi/pending-watcher/releases/download/main/pending-watcher
chmod +x ./pending-watcher
```

### Running
It needs `KUBECONFIG` env var to query the cluster, using default path:
```
KUBECONFIG=~/.kube/config ./pending-watcher
```

### Deploying to Kubernetes

There is a [sample manifest yaml](./deploy-pending-watcher.yaml) with clusteRole and serviceAccount that is needed so that pending-watcher can list pods, you
can use that as a base to deploy it to your cluster to run continuously.

You can review that and then deploy it as is using:
```
kubectl apply -f https://raw.githubusercontent.com/Shadi/pending-watcher/refs/heads/main/deploy-pending-watcher.yaml
```
It will use inCluster config instead of KUBECONFIG env var when its not available.

I originally used this by running it in my machine to quickly finding pending pods and clearing resources on the node
they are pending on, but it I found a use case where this can be helpful to continuously watch and log pods pending in the cluster
so that I can quickly catch logs collector pod being stuck pending on a node, instead of only realizing that the logs collector was pending on the node that had
a pod of that app, so deploying this to the cluster and creating an alert on `Pod pending for more than` log events helped
catching that early.

In some cases the pending-watcher pod can be deployed on that same pod that has no log collector, the quick and dirty
solution is used in the sample manifest where it's deploying 2 replicas with anti-affinity between them to make sure
they are spread on 2 nodes at least, if you think it's not good enough you can deploy even more replicas, deploy it outside
of the cluster, or complicate it more as much as you like.
