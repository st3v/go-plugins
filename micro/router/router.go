// Package router is a micro plugin for defining HTTP routes
package router

import (
	"errors"
	"net/http"

	"github.com/micro/cli"
	"github.com/micro/micro/plugin"
)

type Option func(o *Options)

type router struct {
	opts Options
}

func (r *router) Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "config_source",
			EnvVar: "CONFIG_SOURCE",
			Usage:  "Source to read the config from e.g file, platform",
		},
	}
}

func (r *router) Commands() []cli.Command {
	return nil
}

func (r *router) Handler() plugin.Handler {
	return func(h http.Handler) http.Handler {
		return h
	}
}

func (r *router) Init(ctx *cli.Context) error {
	if c := ctx.String("config_source"); len(c) == 0 && r.opts.Config == nil {
		return errors.New("config source must be defined")
	}
	return nil
}

func (r *router) String() string {
	return "router"
}

func NewRouter(opts ...Option) plugin.Plugin {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	return &router{
		opts: options,
	}
}
