# Kubernetes Registry Plugin for micro
This is a plugin for go-micro that allows you to use Kubernetes as a registry.


## Overview
This registry plugin makes use of Annotations and Labels on a Kubernetes pod
to build a service discovery mechanism.


## Gotchas
* You can only Register/Deregister one node at a time.
* Registering/Deregistering relies on the HOSTNAME Environment Variable, which inside a pod
is the place where it can be retrieved from. (This needs improving)


## Connecting to the Kubernetes API
### Within a pod
If the `--registry_address` flag is omitted, the plugin will securely connect to
the Kubernetes API using the pods "Service Account". No extra configuration is necessary.

Find out more about service accounts here. http://kubernetes.io/docs/user-guide/accessing-the-cluster/

### Outside of Kubernetes
Some functions of the plugin should work, but its not been heavily tested.
Currently no TLS support.
