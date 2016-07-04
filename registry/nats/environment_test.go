package nats_test

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/nats"
)

type environment struct {
	address string

	registryOne   registry.Registry
	registryTwo   registry.Registry
	registryThree registry.Registry

	serviceOne registry.Service
	serviceTwo registry.Service

	nodeOne   registry.Node
	nodeTwo   registry.Node
	nodeThree registry.Node
}

var e environment

func TestMain(m *testing.M) {
	var (
		address = flag.String("url", "nats://localhost:4222", "url to nats server")
	)
	flag.Parse()
	e.address = *address

	e.registryOne = nats.NewRegistry(registry.Addrs(e.address))
	e.registryTwo = nats.NewRegistry(registry.Addrs(e.address))
	e.registryThree = nats.NewRegistry(registry.Addrs(e.address))

	e.serviceOne.Name = "one"
	e.serviceOne.Version = "default"
	e.serviceOne.Nodes = []*registry.Node{&e.nodeOne}

	e.serviceTwo.Name = "two"
	e.serviceTwo.Version = "default"
	e.serviceTwo.Nodes = []*registry.Node{&e.nodeOne, &e.nodeTwo}

	e.nodeOne.Id = "one"
	e.nodeTwo.Id = "two"
	e.nodeThree.Id = "three"

	if err := e.registryOne.Register(&e.serviceOne); err != nil {
		log.Fatal(err)
	}

	if err := e.registryOne.Register(&e.serviceTwo); err != nil {
		log.Fatal(err)
	}

	result := m.Run()

	if err := e.registryOne.Deregister(&e.serviceOne); err != nil {
		log.Fatal(err)
	}

	if err := e.registryOne.Deregister(&e.serviceTwo); err != nil {
		log.Fatal(err)
	}

	os.Exit(result)
}
