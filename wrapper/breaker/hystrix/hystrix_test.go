package hystrix

import (
	"testing"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry/mock"
	"github.com/micro/go-micro/selector"

	"golang.org/x/net/context"
)

func TestBreaker(t *testing.T) {
	// setup
	r := mock.NewRegistry()
	s := selector.NewSelector(selector.Registry(r))

	c := client.NewClient(
		// set the selector
		client.Selector(s),
		// add the breaker wrapper
		client.Wrap(NewClientWrapper()),
	)

	req := c.NewJsonRequest("test.service", "Test.Method", map[string]string{
		"foo": "bar",
	})

	var rsp map[string]interface{}

	// Force to point of trip
	for i := 0; i < 25; i++ {
		c.Call(context.TODO(), req, rsp)
	}

	err := c.Call(context.TODO(), req, rsp)
	if err == nil {
		t.Error("Expecting tripped breaker, got nil error")
	}

	if err.Error() != "hystrix: circuit open" {
		t.Error("Expecting tripped breaker, got %v", err)
	}
}
