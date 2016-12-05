package awsxray

import (
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

type xrayWrapper struct {
	name string
	x    *xray.XRay
	client.Client
}

var (
	TraceHeader = "X-Amzn-Trace-Id"
)

func (x *xrayWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	var err error
	s := getSegment(x.name, ctx)

	defer func() {
		s.HTTP = getHTTP(req, err)
		record(x.x, s)
	}()

	err = x.Client.Call(ctx, req, rsp, opts...)
	return err
}

// NewCallWrapper accepts xray.XRay and returns a Trace Call Wrapper for individual node calls made by the client
func NewCallWrapper(name string, x *xray.XRay) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, addr string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			var err error
			s := getSegment(name, ctx)

			defer func() {
				s.HTTP = getHTTP(req, err)
				record(x, s)
			}()

			err = cf(ctx, addr, req, rsp, opts)
			return err
		}
	}
}

// NewClientWrapper accepts xray.XRay and returns a Trace Client Wrapper which tracks high level service calls
func NewClientWrapper(name string, x *xray.XRay) client.Wrapper {
	return func(c client.Client) client.Client {
		return &xrayWrapper{name, x, c}
	}
}

// NewHandlerWrapper accepts xray.XRay and returns a Trace Handler Wrapper
func NewHandlerWrapper(name string, x *xray.XRay) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			var err error
			s := getSegment(name, ctx)

			defer func() {
				s.HTTP = getHTTP(req, err)
				record(x, s)
			}()

			ctx = context.WithValue(ctx, contextSegmentKey{}, s)
			err = h(ctx, req, rsp)
			return err
		}
	}
}
