package kubernetes

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/micro/go-micro/registry"
)

var data = map[string]string{
	"/api/v1/namespaces/default/endpoints/foo-service":                  `{"kind":"Endpoints","apiVersion":"v1","metadata":{"name":"foo-service","labels":{"name":"foo-service"},"annotations":{}},"subsets":[{"addresses":[{"ip":"10.0.0.1","targetRef":{"kind":"Pod","namespace":"default","name":"foo-service-pod-1"}}],"ports":[{"name":"http","port":80,"protocol":"TCP"}]}]}`,
	"/api/v1/namespaces/default/endpoints/bar-service":                  `{"kind":"Endpoints","apiVersion":"v1","metadata":{"name":"bar-service","labels":{"name":"bar-service"},"annotations":{}},"subsets":[{"addresses":[{"ip":"10.0.1.1","targetRef":{"kind":"Pod","namespace":"default","name":"bar-service-pod-1"}},{"ip":"10.0.1.2","targetRef":{"kind":"Pod","namespace":"default","name":"bar-service-pod-2"}}],"ports":[{"name":"http","port":80,"protocol":"TCP"}]}]}`,
	"/api/v1/namespaces/default/endpoints/baz-service":                  `{"kind":"Endpoints","apiVersion":"v1","metadata":{"name":"baz-service","labels":{"name":"baz-service"},"annotations":{}},"subsets":[{"addresses":[{"ip":"10.0.2.1","targetRef":{"kind":"Pod","namespace":"default","name":"baz-service-pod-1"}},{"ip":"10.0.2.2","targetRef":{"kind":"Pod","namespace":"default","name":"baz-service-pod-2"}},{"ip":"10.0.2.3","targetRef":{"kind":"Pod","namespace":"default","name":"baz-service-pod-3"}}],"ports":[{"name":"http","port":80,"protocol":"TCP"}]}]}`,
	"/api/v1/namespaces/default/pods?labelSelector=micro%3Dfoo-service": `{"kind":"PodList","apiVersion":"v1","metadata":{},"items":[{"metadata":{"name":"foo-service-pod-1","namespace":"default","labels":{"micro":"foo-service"},"annotations":{"micro/endpoints": "[{\"name\":\"foo-service.ep1\"}]","micro/meta":"{\"broker\":\"http\",\"registry\":\"kubernetes\",\"server\":\"rpc\",\"transport\":\"http\"}","micro/name":"foo-service","micro/version":"1"}},"status":{"phase":"Running","podIP":"10.0.0.1"}}]}`,
	"/api/v1/namespaces/default/pods?labelSelector=micro%3Dbar-service": `{"kind":"PodList","apiVersion":"v1","metadata":{},"items":[{"metadata":{"name":"bar-service-pod-1","namespace":"default","labels":{"micro":"bar-service"},"annotations":{"micro/endpoints": "[{\"name\":\"bar-service.ep1\"}]","micro/meta":"{\"broker\":\"http\",\"registry\":\"kubernetes\",\"server\":\"rpc\",\"transport\":\"http\"}","micro/name":"bar-service","micro/version":"1"}},"status":{"phase":"Running","podIP":"10.0.1.1"}},{"metadata":{"name":"bar-service-pod-2","namespace":"default","labels":{"micro":"bar-service"},"annotations":{"micro/endpoints": "[{\"name\":\"bar-service.ep1\"}]", "micro/meta":"{\"broker\":\"http\",\"registry\":\"kubernetes\",\"server\":\"rpc\",\"transport\":\"http\"}","micro/name":"bar-service","micro/version":"1"}},"status":{"phase":"Running","podIP":"10.0.1.2"}}]}`,
	"/api/v1/namespaces/default/pods?labelSelector=micro%3Dbaz-service": `{"kind":"PodList","apiVersion":"v1","metadata":{},"items":[{"metadata":{"name":"baz-service-pod-1","namespace":"default","labels":{"micro":"baz-service"},"annotations":{"micro/endpoints": "[{\"name\":\"baz-service.ep1\"}]","micro/meta":"{\"broker\":\"http\",\"registry\":\"kubernetes\",\"server\":\"rpc\",\"transport\":\"http\"}","micro/name":"baz-service","micro/version":"1"}},"status":{"phase":"Running","podIP":"10.0.2.1"}},{"metadata":{"name":"baz-service-pod-2","namespace":"default","labels":{"micro":"baz-service"},"annotations":{"micro/endpoints": "[{\"name\":\"baz-service.ep1\"}]", "micro/meta":"{\"broker\":\"http\",\"registry\":\"kubernetes\",\"server\":\"rpc\",\"transport\":\"http\"}","micro/name":"baz-service","micro/version":"1"}},"status":{"phase":"Running","podIP":"10.0.2.2"}},{"metadata":{"name":"baz-service-pod-3","namespace":"default","labels":{"micro":"baz-service"},"annotations":{"micro/endpoints": "[{\"name\":\"baz-service.ep1\"},{\"name\":\"baz-service.ep2\"}]","micro/meta":"{\"broker\":\"http\",\"registry\":\"kubernetes\",\"server\":\"rpc\",\"transport\":\"http\"}","micro/name":"baz-service","micro/version":"2"}},"status":{"phase":"Running","podIP":"10.0.2.3"}}]}`,
	"/api/v1/namespaces/default/services":                               `{"kind":"ServiceList","apiVersion":"v1","metadata":{},"items":[{"metadata":{"name":"foo-service","labels":{},"annotations":{}}},{"metadata":{"name":"bar-service","labels":{},"annotations":{}}},{"metadata":{"name":"baz-service","labels":{},"annotations":{}}}]}`,
}

var meta = map[string]string{
	"broker":    "http",
	"registry":  "kubernetes",
	"server":    "rpc",
	"transport": "http",
}

var testdata = map[string][]*registry.Service{
	"foo-service": []*registry.Service{
		{
			Name:    "foo-service",
			Version: "1",
			Nodes: []*registry.Node{
				{
					Id:       "foo-service-pod-1",
					Address:  "10.0.0.1",
					Port:     80,
					Metadata: meta,
				},
			},
			Endpoints: []*registry.Endpoint{
				{
					Name: "foo-service.ep1",
				},
			},
		},
	},
	"bar-service": []*registry.Service{
		{
			Name:    "bar-service",
			Version: "1",
			Nodes: []*registry.Node{
				{
					Id:       "bar-service-pod-1",
					Address:  "10.0.1.1",
					Port:     80,
					Metadata: meta,
				},
				{
					Id:       "bar-service-pod-2",
					Address:  "10.0.1.2",
					Port:     80,
					Metadata: meta,
				},
			},
			Endpoints: []*registry.Endpoint{
				{
					Name: "bar-service.ep1",
				},
			},
		},
	},
	"baz-service": []*registry.Service{
		{
			Name:    "baz-service",
			Version: "1",
			Nodes: []*registry.Node{
				{
					Id:       "baz-service-pod-1",
					Address:  "10.0.2.1",
					Port:     80,
					Metadata: meta,
				},
				{
					Id:       "baz-service-pod-2",
					Address:  "10.0.2.2",
					Port:     80,
					Metadata: meta,
				},
			},
			Endpoints: []*registry.Endpoint{
				{
					Name: "baz-service.ep1",
				},
			},
		}, {
			Name:    "baz-service",
			Version: "2",
			Nodes: []*registry.Node{
				{
					Id:       "baz-service-pod-3",
					Address:  "10.0.2.3",
					Port:     80,
					Metadata: meta,
				},
			},
			Endpoints: []*registry.Endpoint{
				{
					Name: "baz-service.ep1",
				},
				{
					Name: "baz-service.ep2",
				},
			},
		},
	},
}

func hasNodes(a, b []*registry.Node) bool {
	found := 0
	for _, aV := range a {
		for _, bV := range b {
			if reflect.DeepEqual(aV, bV) {
				found++
				break
			}
		}
	}
	return found == len(b)
}

func hasEndpoint(a, b []*registry.Endpoint) bool {
	found := 0
	for _, aV := range a {
		for _, bV := range b {
			if aV.Name == bV.Name {
				found++
				break
			}
		}
	}
	return found == len(b)
}

func hasServices(a, b []*registry.Service) bool {
	found := 0

	for _, aV := range a {
		for _, bV := range b {
			if aV.Name != bV.Name {
				continue
			}
			if aV.Version != bV.Version {
				continue
			}
			if !hasNodes(aV.Nodes, bV.Nodes) {
				continue
			}
			if !hasEndpoint(aV.Endpoints, bV.Endpoints) {
				continue
			}
			found++
			break
		}
	}
	return found == len(b)
}

var defaultHandleFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, data[r.URL.RequestURI()])
})

func TestGetService(t *testing.T) {
	ts := httptest.NewServer(defaultHandleFunc)
	defer ts.Close()

	r := NewRegistry(registry.Addrs(ts.URL))

	fn := func(k string, v []*registry.Service) {
		services, err := r.GetService(k)
		if err != nil {
			t.Errorf("Unexpected error getting service %s: %v", k, err)
		}

		if len(services) != len(v) {
			t.Errorf("Expected %d services for %s, got %d", len(v), k, len(services))
		}

		if !hasServices(services, v) {
			t.Errorf("expected %s to match", k)
		}
	}

	for k, v := range testdata {
		fn(k, v)
	}
}

func TestListServices(t *testing.T) {
	ts := httptest.NewServer(defaultHandleFunc)
	defer ts.Close()

	r := NewRegistry(registry.Addrs(ts.URL))

	svc, err := r.ListServices()
	if err != nil {
		t.Errorf("Unexpected error listing services %v", err)
	}

	var found int
	for _, aV := range svc {
		for _, bV := range testdata[aV.Name] {
			if aV.Name == bV.Name {
				found++
				break
			}
		}
	}

	if found != len(testdata) {
		t.Errorf("did not find all services")
	}

}

func TestRegister(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("did not expect error %v", err)
		}

		b := strings.TrimSpace(string(body))
		expected := `{"metadata":{"annotations":{"micro/endpoints":"[{\"name\":\"foo\",\"request\":{\"name\":\"Req\",\"type\":\"R\",\"values\":null},\"response\":{\"name\":\"Res\",\"type\":\"R\",\"values\":null},\"metadata\":{\"foo\":\"bar\"}}]","micro/meta":"{\"bing\":\"bam\"}","micro/name":"foo-bar-baz","micro/version":"1.5"}}}`

		if b != expected {
			log.Print(string(body))
			t.Fatal("expected json to match")
		}

		w.WriteHeader(http.StatusOK)
	}))

	defer ts.Close()

	r := NewRegistry(registry.Addrs(ts.URL))

	s := &registry.Service{
		Name:    "foo-bar-baz",
		Version: "1.5",
		Endpoints: []*registry.Endpoint{
			{
				Name:     "foo",
				Request:  &registry.Value{Name: "Req", Type: "R"},
				Response: &registry.Value{Name: "Res", Type: "R"},
				Metadata: map[string]string{"foo": "bar"},
			},
		},
		Nodes: []*registry.Node{
			{
				Id:       "foo-service-pod-1",
				Address:  "10.0.0.1",
				Port:     80,
				Metadata: map[string]string{"bing": "bam"},
			},
		},
	}

	err := r.Register(s)
	if err != nil {
		t.Fatalf("did not expect error when registering %v", err)
	}
}

var actions = []string{
	`{"type":"ADDED","object":{"kind":"Endpoints","metadata":{"name":"foo-service"},"subsets":[]}}`,
	`{"type":"MODIFIED","object":{"kind":"Endpoints","metadata":{"name":"foo-service"},"subsets":[{"addresses":[{"ip":"10.0.0.1","targetRef":{"kind":"Pod","namespace":"default","name":"foo-service-pod-1"}}],"ports":[{"name":"http","port":80,"protocol":"TCP"}]}]}}`,
	`{"type":"MODIFIED","object":{"kind":"Endpoints","metadata":{"name":"foo-service"},"subsets":[{"addresses":[{"ip":"10.0.0.1","targetRef":{"kind":"Pod","namespace":"default","name":"foo-service-pod-1"}}, {"ip":"10.0.0.2","targetRef":{"kind":"Pod","namespace":"default","name":"foo-service-pod-2"}}],"ports":[{"name":"http","port":80,"protocol":"TCP"}]}]}}`,
	`{"type":"DELETED","object":{"kind":"Endpoints","metadata":{"name":"foo-service"},"subsets":[{"addresses":[{"ip":"10.0.0.1","targetRef":{"kind":"Pod","namespace":"default","name":"foo-service-pod-1"}}, {"ip":"10.0.0.2","targetRef":{"kind":"Pod","namespace":"default","name":"foo-service-pod-2"}}],"ports":[{"name":"http","port":80,"protocol":"TCP"}]}]}}`,
}

func TestWatch(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected ResponseWriter to be a flusher")
		}

		for _, v := range actions {
			fmt.Fprintf(w, "%s\n", v)
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))

	defer ts.Close()

	r := NewRegistry(registry.Addrs(ts.URL))

	watcher, err := r.Watch()
	if err != nil {
		t.Fatalf("did not expect err %v", err)
	}

	// defer watcher.Stop()
	ch := make(chan *registry.Result, 5)

	go func() {
		for {
			event, err := watcher.Next()
			if err != nil {
				t.Fatalf("did not expect err %v", err)
			}
			ch <- event
		}
	}()

	assert := func(a, b interface{}) {
		if a != b {
			t.Fatalf("expected %s to equal %s", a, b)
		}
	}

	// first
	if e := <-ch; e != nil {
		assert(e.Action, "create")
		assert(e.Service.Name, "foo-service")
		assert(len(e.Service.Nodes), 0)
	}

	// second
	if e := <-ch; e != nil {
		assert(e.Action, "update")
		assert(e.Service.Name, "foo-service")
		assert(len(e.Service.Nodes), 1)
		assert(e.Service.Nodes[0].Address, "10.0.0.1")
	}

	// third
	if e := <-ch; e != nil {
		assert(e.Action, "update")
		assert(e.Service.Name, "foo-service")
		assert(len(e.Service.Nodes), 2)
		assert(e.Service.Nodes[0].Address, "10.0.0.1")
		assert(e.Service.Nodes[1].Address, "10.0.0.2")
	}

	// fourth
	if e := <-ch; e != nil {
		assert(e.Action, "delete")
		assert(e.Service.Name, "foo-service")
		assert(len(e.Service.Nodes), 2)
		assert(e.Service.Nodes[0].Address, "10.0.0.1")
		assert(e.Service.Nodes[1].Address, "10.0.0.2")
	}

}
