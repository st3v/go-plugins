package grpc

import (
	"github.com/micro/go-micro/client"
)

type grpcRequest struct {
	service     string
	method      string
	contentType string
	request     interface{}
	opts        client.RequestOptions
}

func newGRPCRequest(service, method string, request interface{}, contentType string, reqOpts ...client.RequestOption) client.Request {
	var opts client.RequestOptions
	for _, o := range reqOpts {
		o(&opts)
	}

	return &grpcRequest{
		service:     service,
		method:      method,
		request:     request,
		contentType: contentType,
		opts:        opts,
	}
}

func (g *grpcRequest) ContentType() string {
	return g.contentType
}

func (g *grpcRequest) Service() string {
	return g.service
}

func (g *grpcRequest) Method() string {
	return g.method
}

func (g *grpcRequest) Request() interface{} {
	return g.request
}

func (g *grpcRequest) Stream() bool {
	return g.opts.Stream
}
