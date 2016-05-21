# Kubernetes Registry Plugin for micro
This is a plugin for go-micro that allows you to use Kubernetes as a registry.


## Overview
This registry plugin makes use of Kubernetes Services and Endpoints to build a
service discovery mechanism. Endpoints are automatically created by Kubernetes
when a Service is created. Endpoints returns a list of IP and ports for a
service matching the services Label Selector.

When `.Register()` is called, a K8s Service is created if one doesn't already exist.
For each service instance, an annotation is added to the endpoints object. The Label
Selector is added to the labels on the pod the micro service is running in, this
allows the pod to be found by the service.
All this together allows the plugin to build a complete `registry.Service` with all the nodes,
with one call to the K8s API.

## Gotchas
* You can only Register/Deregister one node at a time.
* Registering/Deregistering relies on the HOSTNAME Environment Variable, which inside a pod
is the place where it can be retrieved from. (This needs improving)
* Kubernetes Services wont get removed when a micro service is completed stopped
* This plugin will store annotations on the K8s endpoints object, sometimes they might
hang around, but there is a cleanup process when micro services are registered.
* Kubernetes Services/Endpoints are linked in various ways, for example they
share a name and labels, but they do not share annotations.

## Connecting to the Kubernetes API
### Within a pod
If the `--registry_address` flag is omitted, the plugin will securely connect to
the Kubernetes API using the pods "Service Account". No extra configuration is necessary.

Find out more about service accounts here. http://kubernetes.io/docs/user-guide/accessing-the-cluster/

### Outside of Kubernetes
Some functions of the plugin should work, but its not been heavily tested.
Currently no TLS support.
