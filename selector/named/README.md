# Named selector

The named selector returns the named service as a node for every request. This is useful where you want to 
offload discovery and balancing too a message bus.

When a service uses a message bus such as NATS for transport it will call `transport.Listen` with it's service name. 
This will force any instance of the service to subscribe to a topic with its service name. In combination with 
the named selector, we'll offload loadbalancing to the message bus itself.

## Usage

```go
selector := named.NewSelector()

service := micro.NewService(
	client.NewClient(client.Selector(selector))
)
```
