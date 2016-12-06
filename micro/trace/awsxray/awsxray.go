// Package awsxray is a micro plugin for whitelisting service requests
package awsxray

import (
	"net/http"

	"github.com/micro/cli"
	"github.com/micro/go-micro/client"
	xray "github.com/micro/go-plugins/wrapper/trace/awsxray"
	"github.com/micro/micro/plugin"
)

type awsxray struct {
	opts Options
	rec  recorder
}

var (
	TraceHeader = "X-Amzn-Trace-Id"
)

func (x *awsxray) Flags() []cli.Flag {
	return nil
}

func (x *awsxray) Commands() []cli.Command {
	return nil
}

func (x *awsxray) Handler() plugin.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := newSegment(x.opts.Name, r)
			// use our own writer
			xw := &writer{w, 200}
			// serve request
			h.ServeHTTP(xw, r)
			// set status
			s.SetStatus(xw.status)
			// send segment asynchronously
			go x.rec.record(s)
		})
	}
}

func (x *awsxray) Init(ctx *cli.Context) error {
	opts := []xray.Option{
		xray.WithName(x.opts.Name),
		xray.WithClient(x.opts.Client),
		xray.WithDaemon(x.opts.Daemon),
	}

	// setup client
	c := client.DefaultClient
	c = xray.NewClientWrapper(opts...)(c)
	c.Init(client.WrapCall(xray.NewCallWrapper(opts...)))
	client.DefaultClient = c
	return nil
}

func (x *awsxray) String() string {
	return "awsxray"
}

func NewXRayPlugin(opts ...Option) plugin.Plugin {
	options := Options{
		Name:   "go.micro.http",
		Daemon: "localhost:2000",
	}

	for _, o := range opts {
		o(&options)
	}

	return &awsxray{
		opts: options,
		rec:  recorder{options},
	}
}
