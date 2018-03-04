// Package quota implements a client side wrapper for rate limiting using the quota-srv
package quota

import (
	"fmt"
	"hash/crc32"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"

	proto "github.com/micro/quota-srv/proto"

	"context"
)

type quota struct {
	service string
	qcl     proto.QuotaClient
	client.Client
}

func (q *quota) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	// unique key for this resource
	key := fmt.Sprintf("%s:%s:%s", q.service, req.Service(), req.Method())

	// checksum it
	cs := crc32.ChecksumIEEE([]byte(key))

	// shard the request
	nOpts := append(opts, client.WithSelectOption(
		// create a selector strategy
		selector.WithStrategy(func(services []*registry.Service) selector.Next {
			// flatten
			var nodes []*registry.Node
			for _, service := range services {
				nodes = append(nodes, service.Nodes...)
			}

			// create the next func that always returns our node
			return func() (*registry.Node, error) {
				if len(nodes) == 0 {
					return nil, selector.ErrNoneAvailable
				}
				return nodes[cs%uint32(len(nodes))], nil
			}
		}),
	))

	// get quota
	qrsp, err := q.qcl.Allocate(ctx, &proto.AllocateRequest{
		Resource:   fmt.Sprintf("%s.%s", req.Service(), req.Method()),
		Bucket:     q.service,
		Allocation: 1,
	}, nOpts...)
	if err != nil {
		return err
	}

	switch qrsp.Status {
	// status ok. call service
	case proto.AllocateResponse_OK:
		return q.Client.Call(ctx, req, rsp, opts...)
	// rate limited
	case proto.AllocateResponse_REJECT_TOO_MANY_REQUESTS:
		return errors.New(
			"go.micro.srv.quota",
			proto.AllocateResponse_REJECT_TOO_MANY_REQUESTS.String(),
			429,
		)
	// internal error
	default:
		return errors.InternalServerError("go.micro.srv.quota", qrsp.Status.String())
	}
}

// NewClientWrapper is a wrapper which calls uses the quota-srv for rate limiting
// Key argument should be a unique key for your service
func NewClientWrapper(service string) client.Wrapper {
	return func(c client.Client) client.Client {
		return &quota{
			service: service,
			qcl:     proto.NewQuotaClient("go.micro.srv.quota", c),
			Client:  c,
		}
	}
}
