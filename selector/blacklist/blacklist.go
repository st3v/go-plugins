package blacklist

import (
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/selector/blacklist"
)

/*
	Blacklist selector is a client side load balancer for go-micro.
	It uses random load balancing to balance requests across services.
	When nodes are Marked with an error it will blacklist them for 60 seconds.
	Implementation here https://godoc.org/github.com/micro/go-micro/selector/blacklist
	We add a link here for completeness
*/

func NewSelector(opts ...selector.Option) selector.Selector {
	return blacklist.NewSelector(opts...)
}
