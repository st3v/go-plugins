package memory

import (
	"github.com/micro/go-micro/registry"

	"golang.org/x/net/context"
)

type contextSecretKeyT string

var (
	contextSecretKey = contextSecretKeyT("github.com/micro/go-plugins/registry/memory")
)

func SecretKey(k []byte) registry.Option {
	return func(o *registry.Options) {
		o.Context = context.WithValue(o.Context, contextSecretKey, k)
	}
}
