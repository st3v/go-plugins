package mqtt

import (
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/broker/mqtt"
)

/*
	MQTT Broker is an mqtt based broker that can be used for IoT.
	Find the implementation at mqtts://godoc.org/github.com/micro/go-micro/broker/mqtt.
	We add a link here for completeness
*/

func NewBroker(opts ...broker.Option) broker.Broker {
	return mqtt.NewBroker(opts...)
}
