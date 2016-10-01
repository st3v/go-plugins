package googlepubsub

import (
	"github.com/micro/go-micro/broker"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

type clientOptionKey struct{}

type projectIDKey struct{}

// ClientOption is a broker Option which allows google pubsub client options to be
// set for the client
func ClientOption(c ...option.ClientOption) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, clientOptionKey{}, c)
	}
}

// ProjectID provides an option which sets the google project id
func ProjectID(id string) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, projectIDKey{}, id)
	}
}
