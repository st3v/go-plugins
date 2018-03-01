# Datadog Metrics

This is a [go-os/metrics](https://github.com/micro/go-os/tree/master/metrics) plugin for datadog.
It pushes tagged statsd style metrics to datadog.

For middleware examples see [andrew-jones/go-micro-middleware](https://github.com/andrew-jones/go-micro-middleware).

**Note:** Counters and gauges are untested as my middleware only creates histograms at this stage.

## Usage

```go
func main() {
	// Create datadog metrics collector
	m := datadog.NewMetrics(
		metrics.Namespace("micro"),
		metrics.WithFields(metrics.Fields{
			"service": "greeter",
		}),
		metrics.Collectors("dd-agent:8125"),
	)
	defer m.Close()

	// Create broker
	b := rabbitmq.NewBroker(
		broker.Addrs("amqp://guest:guest@rabbit:5672"),
	)
	if err := b.Init(); err != nil {
		log.Fatalf("Unexpected init error: %v", err)
	}
	if err := b.Connect(); err != nil {
		log.Fatalf("Unexpected connect error: %v", err)
	}
	// Wrap the broker in logging and metric middleware
	b = middleware.LogBrokerWrapper(
		middleware.MetricBrokerWrapper(b, m, time.Millisecond),
	)
    // Subscribe to broker, setting up a durable queue with auto ack disabled
    // eventsubscriber is a package implementing broker.Handler
    _, err := b.Subscribe(
		"routing.key",
		eventsubscriber.NewBrokerHandler(subscriber.NewSubscriber("greeter")),
		broker.Queue("greeter"),
		broker.DisableAutoAck(),
		rabbitmq.DurableQueue(),
	)
	if err != nil {
		log.Fatalf("Could not subscribe handler to broker: %v", err)
	}

	// Create service
	service := micro.NewService(
		micro.Name("greeter"),
		micro.Broker(b),
		micro.Server(
			server.NewServer(
				server.Name("greeter"),
				server.WrapHandler(middleware.MetricHandlerWrapper(m, time.Millisecond)),
			),
		),
	)
	// setup command line usage
	service.Init()
	// Run server
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
```
