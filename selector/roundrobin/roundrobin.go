package roundrobin

import (
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/selector/roundrobin"
)

/*
	RoundRobin selector is a client side load balancer for go-micro.
	It uses round robin load balancing to balance requests across services.
	Implementation here https://godoc.org/github.com/micro/go-micro/selector/roundrobin
	We add a link here for completeness
*/

func NewSelector(opts ...selector.Option) selector.Selector {
	return roundrobin.NewSelector(opts...)
}
