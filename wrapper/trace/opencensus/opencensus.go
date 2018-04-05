// Package opencensus provides wrappers for OpenCensus tracing.
package opencensus

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/micro/go-log"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

const (
	// TracePropagationField is the key for the tracing context
	// that will be injected in go-micro's metadata.
	TracePropagationField = "X-Trace-Context"
)

// clientWrapper wraps an RPC client and adds tracing.
type clientWrapper struct {
	client.Client
}

func injectTraceIntoCtx(ctx context.Context, span *trace.Span) context.Context {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}

	spanCtx := propagation.Binary(span.SpanContext())
	md[TracePropagationField] = base64.RawStdEncoding.EncodeToString(spanCtx)

	return metadata.NewContext(ctx, md)
}

// Call implements client.Client.Call.
func (w *clientWrapper) Call(
	ctx context.Context,
	req client.Request,
	rsp interface{},
	opts ...client.CallOption) error {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("rpc/call/%s/%s", req.Service(), req.Method()))
	defer span.End()

	ctx = injectTraceIntoCtx(ctx, span)

	return w.Client.Call(ctx, req, rsp, opts...)
}

// Publish implements client.Client.Publish.
func (w *clientWrapper) Publish(ctx context.Context, p client.Publication, opts ...client.PublishOption) error {
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("rpc/publish/%s", p.Topic()))
	defer span.End()

	ctx = injectTraceIntoCtx(ctx, span)

	return w.Client.Publish(ctx, p, opts...)
}

// NewClientWrapper returns a client.Wrapper
// that adds monitoring to outgoing requests.
func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		return &clientWrapper{c}
	}
}

func getTraceFromCtx(ctx context.Context) *trace.SpanContext {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}

	encodedTraceCtx, ok := md[TracePropagationField]
	if !ok {
		log.Log("Missing trace context in incoming request")
		return nil
	}

	traceCtxBytes, err := base64.RawStdEncoding.DecodeString(encodedTraceCtx)
	if err != nil {
		log.Logf("Could not decode trace context: %s", err.Error())
		return nil
	}

	spanCtx, ok := propagation.FromBinary(traceCtxBytes)
	if !ok {
		log.Log("Could not decode trace context from binary")
		return nil
	}

	return &spanCtx
}

// NewHandlerWrapper returns a server.HandlerWrapper
// that adds tracing to incoming requests.
func NewHandlerWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			var span *trace.Span
			defer span.End()

			spanCtx := getTraceFromCtx(ctx)
			if spanCtx != nil {
				span = trace.NewSpanWithRemoteParent(
					fmt.Sprintf("rpc/handle/%s/%s", req.Service(), req.Method()),
					*spanCtx,
					trace.StartOptions{},
				)
				ctx = trace.WithSpan(ctx, span)
			} else {
				ctx, span = trace.StartSpan(
					ctx,
					fmt.Sprintf("rpc/handle/%s/%s", req.Service(), req.Method()),
				)
			}

			return fn(ctx, req, rsp)
		}
	}
}

// NewSubscriberWrapper returns a server.SubscriberWrapper
// that adds tracing to subscription requests.
func NewSubscriberWrapper() server.SubscriberWrapper {
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, p server.Publication) error {
			var span *trace.Span
			defer span.End()

			spanCtx := getTraceFromCtx(ctx)
			if spanCtx != nil {
				span = trace.NewSpanWithRemoteParent(
					fmt.Sprintf("rpc/subscribe/%s", p.Topic()),
					*spanCtx,
					trace.StartOptions{},
				)
				ctx = trace.WithSpan(ctx, span)
			} else {
				ctx, span = trace.StartSpan(
					ctx,
					fmt.Sprintf("rpc/subscribe/%s", p.Topic()),
				)
			}

			return fn(ctx, p)
		}
	}
}
