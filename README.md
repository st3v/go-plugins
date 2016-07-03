# Go Plugins [![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/micro/go-plugins?status.svg)](https://godoc.org/github.com/micro/go-plugins) [![Travis CI](https://travis-ci.org/micro/go-plugins.svg?branch=master)](https://travis-ci.org/micro/go-plugins) [![Go Report Card](https://goreportcard.com/badge/micro/go-plugins)](https://goreportcard.com/report/github.com/micro/go-plugins)

A repository for go-micro and go-platform plugins. The go-micro framework is a pluggable architecture, here we provide additional plugins to 
swap out the defaults. 

Check out the Micro on NATS blog post to learn more about plugins [https://blog.micro.mu/2016/04/11/micro-on-nats.html](https://blog.micro.mu/2016/04/11/micro-on-nats.html).

Contributions welcome! Join the community to discuss further.

- [Mailing List](https://groups.google.com/forum/#!forum/microhq) 
- [Slack](https://micro-services.slack.com) : [auto-invite](http://micro-invites.herokuapp.com/)

## What's here?

Directory	|	Description
---		|	---
Broker		|	Asynchronous Pub/Sub; NATS, NSQ, RabbitMQ, Kafka	
Codec		|	RPC Encoding; BSON, Mercury
Registry	|	Service Discovery; Etcd, Gossip, NATS
Selector	|	Node Selection; Label, Mercury
Transport	|	Synchronous Request/Response; NATS, RabbitMQ
Wrappers	|	Client/Server middleware; Circuit Breakers, Rate Limit

## Community Contributions

Feature		|	Description		|	Author
----------	|	------------		|	--------
[Registry/Kubernetes](https://godoc.org/github.com/micro/go-plugins/registry/kubernetes)	|	Service discovery via the Kubernetes API	|	[@nickjackson](https://github.com/nickjackson)
[Registry/Zookeeper](https://godoc.org/github.com/micro/go-plugins/registry/zookeeper)	|	Service discovery using Zookeeper	|	[@HeavyHorst](https://github.com/HeavyHorst)

## Usage

Plugins can be added to go-micro in the following ways. By doing so they'll be available to set via command line args or environment variables.

```go
import (
	"github.com/micro/go-micro/cmd"
	_ "github.com/micro/go-plugins/broker/rabbitmq"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	_ "github.com/micro/go-plugins/transport/nats"
)

func main() {
	cmd.Init()
}
```

OR use them directly

```go
import (
	"github.com/micro/go-plugins/registry/kubernetes"
)

func main() {
	r := kubernetes.NewRegistry([]string{}) // default to using env vars for master API
}
```
