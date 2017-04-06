package rabbitmq

import (
	"github.com/micro/go-micro/broker"
	"golang.org/x/net/context"
)

type durableQueueKey struct{}
type exchangeKey struct{}

// DurableQueue creates a durable queue when subscribing.
func DurableQueue() broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, durableQueueKey{}, true)
	}
}

// Exchange is an option to set the Exchange
func Exchange(e string) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, exchangeKey{}, e)
	}
}
