package mercury

import (
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
)

type mercurySelector struct{}

func (r *mercurySelector) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	node := &registry.Node{
		Id:      service,
		Address: service,
	}

	return func() (*registry.Node, error) {
		return node, nil
	}, nil
}

func (r *mercurySelector) Mark(service string, node *registry.Node, err error) {
	return
}

func (r *mercurySelector) Reset(service string) {
	return
}

func (r *mercurySelector) Close() error {
	return nil
}

func (r *mercurySelector) String() string {
	return "mercury"
}

func NewSelector(opts ...selector.Option) selector.Selector {
	return &mercurySelector{}
}
