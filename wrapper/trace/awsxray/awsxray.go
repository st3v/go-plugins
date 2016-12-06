package awsxray

import (
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

type xrayWrapper struct {
	opts Options
	r    recorder
	client.Client
}

var (
	TraceHeader = "X-Amzn-Trace-Id"
)

func (x *xrayWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	var err error
	s := getSegment(x.opts.Name, ctx)

	defer func() {
		s.HTTP = getHTTP(req.Service(), req.Method(), err)
		s.SetStatus(err)
		go x.r.record(s)
	}()

	ctx = newContext(ctx, s)
	err = x.Client.Call(ctx, req, rsp, opts...)
	return err
}

// NewCallWrapper accepts Options and returns a Trace Call Wrapper for individual node calls made by the client
func NewCallWrapper(opts ...Option) client.CallWrapper {
	options := Options{
		Name:   "go.micro.client.CallFunc",
		Daemon: "localhost:2000",
	}

	for _, o := range opts {
		o(&options)
	}

	r := recorder{options}

	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, addr string, req client.Request, rsp interface{}, opts client.CallOptions) error {
			var err error
			s := getSegment(options.Name, ctx)

			defer func() {
				s.HTTP = getHTTP(addr, req.Method(), err)
				s.SetStatus(err)
				go r.record(s)
			}()

			ctx = newContext(ctx, s)
			err = cf(ctx, addr, req, rsp, opts)
			return err
		}
	}
}

// NewClientWrapper accepts Options and returns a Trace Client Wrapper which tracks high level service calls
func NewClientWrapper(opts ...Option) client.Wrapper {
	options := Options{
		Name:   "go.micro.client.Call",
		Daemon: "localhost:2000",
	}

	for _, o := range opts {
		o(&options)
	}

	return func(c client.Client) client.Client {
		return &xrayWrapper{options, recorder{options}, c}
	}
}

// NewHandlerWrapper accepts Options and returns a Trace Handler Wrapper
func NewHandlerWrapper(opts ...Option) server.HandlerWrapper {
	options := Options{
		Daemon: "localhost:2000",
	}

	for _, o := range opts {
		o(&options)
	}

	r := recorder{options}

	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := options.Name
			if len(name) == 0 {
				// default name
				name = req.Service() + "." + req.Method()
			}

			var err error
			s := getSegment(name, ctx)

			defer func() {
				s.HTTP = getHTTP(req.Service(), req.Method(), err)
				s.SetStatus(err)
				go r.record(s)
			}()

			ctx = newContext(ctx, s)
			err = h(ctx, req, rsp)
			return err
		}
	}
}
