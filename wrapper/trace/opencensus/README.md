# OpenCensus wrappers

OpenCensus wrappers propagate traces (spans) accross services.

## Usage

```go
service := micro.NewService(
    micro.Name("go.micro.srv.greeter"),
    micro.WrapClient(opencensus.NewClientWrapper()),
    micro.WrapHandler(opencensus.NewHandlerWrapper()),
    micro.WrapSubscriber(opencensus.NewSubscriberWrapper()),
)
```
