# Kubernetes Registry Plugin for micro
This is a plugin for go-micro that allows you to use Kubernetes as a registry.


## Overview
You must have a "Service" and "Replication Controller" (or "Deployment") manifest
for each service you deploy for the plugin to function.

You can use the notion of "Replication Controllers" (or "Deployments") to scale
your microservice. Make sure to include the label `micro: "<service-name>"` on
your pod specification.

The plugin makes use of Kubernetes "Services" to provide a list of addresses for
each instance/pod of your service.


## Connecting to the Kubernetes API
### Within a pod
If the `--registry_address` flag is omitted, the plugin will securely connect to
the Kubernetes API using the pods "Service Account". No extra configuration is necessary.

Find out more about service accounts here. http://kubernetes.io/docs/user-guide/accessing-the-cluster/

### Outside of Kubernetes
Some functions of the plugin should work, but its not been heavily tested.
Currently no TLS support.
