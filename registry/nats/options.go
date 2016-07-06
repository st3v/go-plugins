package nats

import (
	"github.com/micro/go-micro/registry"
	"golang.org/x/net/context"
)

const quorumKey = "quorum"

var DefaultQuorum = 0

func Quorum(n int) registry.Option {
	return func(o *registry.Options) {
		o.Context = context.WithValue(o.Context, quorumKey, n)
	}
}

func GetQuorum(o *registry.Options) int {
	if o.Context == nil {
		return DefaultQuorum
	}

	value := o.Context.Value(quorumKey)
	if v, ok := value.(int); ok {
		return v
	} else {
		return DefaultQuorum
	}
}
