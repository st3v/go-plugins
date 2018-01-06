package nats

import (
	"github.com/micro/go-micro/transport"
	"github.com/nats-io/nats"
)

var (
	DefaultNatsOptions = nats.GetDefaultOptions()

	optionsKey = optionsKeyType{}
)

type optionsKeyType struct{}

type transportOptions struct {
	natsOptions nats.Options
}

// NatsOptions allow to inject a nats.Options struct for configuring
// the nats connection
func NatsOptions(nopts nats.Options) transport.Option {
	return func(o *transport.Options) {
		no := o.Context.Value(optionsKey).(*transportOptions)
		no.natsOptions = nopts
	}
}
