package random

import (
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/selector/random"
)

/*
	Random selector is a client side load balancer for go-micro.
	It uses random hashed load balancing to balance requests across services.
	Implementation here https://godoc.org/github.com/micro/go-micro/selector/random
	We add a link here for completeness
*/

func NewSelector(opts ...selector.Option) selector.Selector {
	return random.NewSelector(opts...)
}
