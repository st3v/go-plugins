// Package opentracing provides wrappers for OpenTracing
package opentracing

import (
	"fmt"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
)

type otWrapper struct {
	ot opentracing.Tracer
	client.Client
}

func traceIntoContext(ctx context.Context, tracer opentracing.Tracer, name string) (context.Context, error) {
	md, _ := metadata.FromContext(ctx)
	var sp opentracing.Span
	wireContext, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(md))
	if err != nil {
		sp = tracer.StartSpan(name)
	} else {
		sp = tracer.StartSpan(name, opentracing.ChildOf(wireContext))
	}
	defer sp.Finish()
	if err := sp.Tracer().Inject(sp.Context(), opentracing.TextMap, opentracing.TextMapCarrier(md)); err != nil {
		return nil, err
	}
	ctx = metadata.NewContext(ctx, md)
	return ctx, nil
}

func (o *otWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Method())
	ctx, err := traceIntoContext(ctx, o.ot, name)
	if err != nil {
		return err
	}
	return o.Client.Call(ctx, req, rsp, opts...)
}

func (o *otWrapper) Publish(ctx context.Context, p client.Publication, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	ctx, err := traceIntoContext(ctx, o.ot, name)
	if err != nil {
		return err
	}
	return o.Client.Publish(ctx, p, opts...)
}

// NewClientWrapper accepts an open tracing Trace and returns a Client Wrapper
func NewClientWrapper(ot opentracing.Tracer) client.Wrapper {
	return func(c client.Client) client.Client {
		return &otWrapper{ot, c}
	}
}

// NewHandlerWrapper accepts an opentracing Tracer and returns a Handler Wrapper
func NewHandlerWrapper(ot opentracing.Tracer) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := fmt.Sprintf("%s.%s", req.Service(), req.Method())
			ctx, err := traceIntoContext(ctx, ot, name)
			if err != nil {
				return err
			}
			return h(ctx, req, rsp)
		}
	}
}

// NewSubscriberWrapper accepts an opentracing Tracer and returns a Subscriber Wrapper
func NewSubscriberWrapper(ot opentracing.Tracer) server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Publication) error {
			name := "Pub to " + msg.Topic()
			ctx, err := traceIntoContext(ctx, ot, name)
			if err != nil {
				return err
			}
			return next(ctx, msg)
		}
	}
}
