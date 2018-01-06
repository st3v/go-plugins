package nats

import (
	"github.com/micro/go-micro/registry"
	"github.com/nats-io/nats"
	"golang.org/x/net/context"
)

type contextQuorumKey struct{}
type optionsKeyType struct{}

var (
	DefaultQuorum      = 0
	DefaultNatsOptions = nats.GetDefaultOptions()

	optionsKey = optionsKeyType{}
)

type registryOptions struct {
	natsOptions nats.Options
	queryTopic  string
	watchTopic  string
}

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

// NatsOptions allow to inject a nats.Options struct for configuring
// the nats connection
func NatsOptions(nopts nats.Options) registry.Option {
	return func(o *registry.Options) {
		no := o.Context.Value(optionsKey).(*registryOptions)
		no.natsOptions = nopts
	}
}

// QueryTopic allows to set a custom nats topic on which service registries
// query (survey) other services. All registries listen on this topic and
// then respond to the query message.
func QueryTopic(s string) registry.Option {
	return func(o *registry.Options) {
		no := o.Context.Value(optionsKey).(*registryOptions)
		no.queryTopic = s
	}
}

// WatchTopic allows to set a custom nats topic on which registries broadcast
// changes (e.g. when services are added, updated or removed). Since we don't
// have a central registry service, each service typically broadcasts in a
// determined frequency on this topic.
func WatchTopic(s string) registry.Option {
	return func(o *registry.Options) {
		no := o.Context.Value(optionsKey).(*registryOptions)
		no.watchTopic = s
	}
}
