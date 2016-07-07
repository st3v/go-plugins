package nats

import (
	"github.com/micro/go-micro/registry"
	"golang.org/x/net/context"
)

type contextQuorumKey struct{}

var DefaultQuorum = 0

func Quorum(n int) registry.Option {
	return func(o *registry.Options) {
		o.Context = context.WithValue(o.Context, contextQuorumKey{}, n)
	}
}

func getQuorum(o registry.Options) int {
	if o.Context == nil {
		return DefaultQuorum
	}

	value := o.Context.Value(contextQuorumKey{})
	if v, ok := value.(int); ok {
		return v
	} else {
		return DefaultQuorum
	}
}
