package server

import (
	"github.com/micro/go-plugins/Godeps/_workspace/src/github.com/micro/go-micro/server/proto/health"
	"github.com/micro/go-plugins/Godeps/_workspace/src/golang.org/x/net/context"
)

type Debug struct{}

func (d *Debug) Health(ctx context.Context, req *health.Request, rsp *health.Response) error {
	rsp.Status = "ok"
	return nil
}

func registerHealthChecker(s Server) {
	s.Handle(s.NewHandler(&Debug{}))
}
