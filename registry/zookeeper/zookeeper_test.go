package zookeeper

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/micro/go-micro/registry"
	"github.com/samuel/go-zookeeper/zk"
)

func TestZookeeper(t *testing.T) {
	testData := map[string][]*registry.Service{
		"foo": []*registry.Service{
			{
				Name:    "foo",
				Version: "1.0.0",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.0-123",
						Address: "localhost",
						Port:    9999,
					},
					{
						Id:      "foo-1.0.0-321",
						Address: "localhost",
						Port:    9999,
					},
				},
			},
			{
				Name:    "foo",
				Version: "1.0.1",
				Nodes: []*registry.Node{
					{
						Id:      "foo-1.0.1-321",
						Address: "localhost",
						Port:    6666,
					},
				},
			},
		},
		"bar": []*registry.Service{
			{
				Name:    "bar",
				Version: "default",
				Nodes: []*registry.Node{
					{
						Id:      "bar-1.0.0-123",
						Address: "localhost",
						Port:    9999,
					},
					{
						Id:      "bar-1.0.0-321",
						Address: "localhost",
						Port:    9999,
					},
				},
			},
			{
				Name:    "bar",
				Version: "latest",
				Nodes: []*registry.Node{
					{
						Id:      "bar-1.0.1-321",
						Address: "localhost",
						Port:    6666,
					},
				},
			},
		},
	}

	c, err := zk.StartTestCluster(1, ioutil.Discard, ioutil.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Stop()

	var addrs []string

	for _, srv := range c.Servers {
		addrs = append(addrs, fmt.Sprintf("%s:%d", "127.0.0.1", srv.Port))
	}

	r := NewRegistry(registry.Addrs(addrs...))

	for _, services := range testData {
		for _, service := range services {
			// register service
			if err := r.Register(service); err != nil {
				t.Fatal(err)
			}
		}
	}

	for name, services := range testData {
		srvs, err := r.GetService(name)
		if err != nil {
			t.Fatal(err)
		}

		for _, serv := range services {
			var seen bool
			for _, srv := range srvs {
				if serv.Name != srv.Name {
					continue
				}

				if serv.Version != srv.Version {
					continue
				}

				seen = true

				for _, snode := range serv.Nodes {
					var found bool

					for _, node := range srv.Nodes {
						if snode.Id != node.Id {
							continue
						}

						if snode.Address != node.Address {
							continue
						}

						if snode.Port != node.Port {
							continue
						}

						found = true
					}

					if !found {
						t.Fatalf("%+v not found: got %+v", snode, srvs)
					}
				}
			}
			if !seen {
				t.Fatalf("%+v not found: got %+v", serv, srvs)
			}
		}
	}

	for _, services := range testData {
		for _, service := range services {
			// deregister service
			if err := r.Deregister(service); err != nil {
				t.Fatal(err)
			}
		}
	}
}
