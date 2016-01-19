package nsq

import (
	"github.com/micro/go-micro/broker"

	"golang.org/x/net/context"
)

type contextKeyT string

var (
	concurrentHandlerKey = contextKeyT("github.com/micro/go-plugins/broker/nsq/concurrentHandlers")
)

func ConcurrentHandlers(n int) broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		o.Context = context.WithValue(o.Context, concurrentHandlerKey, n)
	}
}
