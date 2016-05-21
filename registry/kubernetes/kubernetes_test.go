package kubernetes

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/selector/cache"
	"github.com/micro/go-plugins/registry/kubernetes/client"
	"github.com/micro/go-plugins/registry/kubernetes/client/mock"
	"github.com/micro/go-plugins/registry/kubernetes/client/watch"
)

var meta = map[string]string{
	"broker":    "http",
	"registry":  "kubernetes",
	"server":    "rpc",
	"transport": "http",
}

var testdata = map[string][]*registry.Service{
	"foo.service": []*registry.Service{
		{
			Name:    "foo.service",
			Version: "1",
			Nodes: []*registry.Node{
				{
					Id:      "foo-service-1",
					Address: "10.0.0.1",
					Port:    80,
					Metadata: map[string]string{
						"foo": "bar",
					},
				},
			},
			Endpoints: []*registry.Endpoint{
				{
					Name: "foo-service.ep1",
				},
			},
		},
		{
			Name:    "foo.service",
			Version: "1",
			Nodes: []*registry.Node{
				{
					Id:      "foo-service-2",
					Address: "10.0.0.2",
					Port:    80,
					Metadata: map[string]string{
						"v": "1",
					},
				},
			},
			Endpoints: []*registry.Endpoint{
				{
					Name: "foo-service.ep1",
				},
				{
					Name: "foo-service.ep2",
				},
			},
		},
	},
	"bar.service": []*registry.Service{
		{
			Name:    "bar.service",
			Version: "1",
			Nodes: []*registry.Node{
				{
					Id:       "bar-service-pod-1",
					Address:  "10.0.0.1",
					Port:     81,
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
}

var (
	mockClient = mock.NewClient()
)

func mockK8s() {
	mockClient.Pods = map[string]*client.Pod{
		"pod-1": &client.Pod{
			Metadata: &client.Meta{
				Name:        "pod-1",
				Labels:      make(map[string]*string),
				Annotations: make(map[string]*string),
			},
			Status: &client.Status{
				PodIP: "10.0.0.1",
			},
		},
		"pod-2": &client.Pod{
			Metadata: &client.Meta{
				Name:        "pod-2",
				Labels:      make(map[string]*string),
				Annotations: make(map[string]*string),
			},
			Status: &client.Status{
				PodIP: "10.0.0.2",
			},
		},
	}
}

func teardownK8s() {
	mock.Teardown(mockClient)
}

func newMockRegistry(opts ...registry.Option) registry.Registry {
	return &kregistry{
		client:  mockClient,
		timeout: time.Second * 1,
	}
}

func TestRegister(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	os.Setenv("HOSTNAME", "pod-1")
	svc := testdata["foo.service"][0]

	r := newMockRegistry()

	if err := r.Register(svc); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}

	// check pod has correct labels/annotations
	p := mockClient.Pods["pod-1"]
	svcLabel, ok := p.Metadata.Labels[svcSelectorPrefix+"foo.service"]
	if !ok || *svcLabel != svcSelectorValue {
		t.Fatalf("expected to have pod selector label")
	}

	// check k8s service has expected result
	s, ok := mockClient.Endpoints["foo-service"]
	if !ok {
		t.Fatalf("expected service to exist")
	}
	svcData, ok := s.Metadata.Annotations[annotationServiceKeyPrefix+"pod-1"]
	if !ok || len(*svcData) == 0 {
		t.Fatalf("expected to have annotation")
	}

	// unmarshal service data from annotation and compare
	// service passed in to .Register()
	var service *registry.Service
	if err := json.Unmarshal([]byte(*svcData), &service); err != nil {
		t.Fatalf("did not expect register unmarshal to fail %v", err)
	}
	if !reflect.DeepEqual(svc, service) {
		t.Fatal("services did not match")
	}
}

func TestRegisterTwoDifferentServicesOnePod(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	os.Setenv("HOSTNAME", "pod-1")
	svc1 := testdata["foo.service"][0]
	svc2 := testdata["bar.service"][0]

	r := newMockRegistry()

	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}
	if err := r.Register(svc2); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}

	// check pod has correct labels/annotations
	p := mockClient.Pods["pod-1"]
	if svcLabel1, ok := p.Metadata.Labels[svcSelectorPrefix+"foo.service"]; !ok || *svcLabel1 != svcSelectorValue {
		t.Fatalf("expected to have pod selector label for service one")
	}
	if svcLabel2, ok := p.Metadata.Labels[svcSelectorPrefix+"bar.service"]; !ok || *svcLabel2 != svcSelectorValue {
		t.Fatalf("expected to have pod selector label for service two")
	}

	// check k8s service one has expected result
	s1, ok := mockClient.Endpoints["foo-service"]
	if !ok {
		t.Fatalf("expected service to exist")
	}
	svcData1, ok := s1.Metadata.Annotations[annotationServiceKeyPrefix+"pod-1"]
	if !ok || len(*svcData1) == 0 {
		t.Fatalf("expected to have annotation")
	}

	// check k8s service one has expected result
	s2, ok := mockClient.Endpoints["bar-service"]
	if !ok {
		t.Fatalf("expected service to exist")
	}
	svcData2, ok := s2.Metadata.Annotations[annotationServiceKeyPrefix+"pod-1"]
	if !ok || len(*svcData2) == 0 {
		t.Fatalf("expected to have annotation")
	}

	// unmarshal service data from annotation and compare
	// service passed in to .Register()
	var service1 *registry.Service
	if err := json.Unmarshal([]byte(*svcData1), &service1); err != nil {
		t.Fatalf("did not expect register unmarshal to fail %v", err)
	}
	if !reflect.DeepEqual(svc1, service1) {
		t.Fatal("services did not match")
	}

	var service2 *registry.Service
	if err := json.Unmarshal([]byte(*svcData2), &service2); err != nil {
		t.Fatalf("did not expect register unmarshal to fail %v", err)
	}
	if !reflect.DeepEqual(svc2, service2) {
		t.Fatal("services did not match")
	}
}

func TestRegisterTwoDifferentServicesTwoPods(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	svc1 := testdata["foo.service"][0]
	svc2 := testdata["bar.service"][0]

	r := newMockRegistry()

	os.Setenv("HOSTNAME", "pod-1")
	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}
	os.Setenv("HOSTNAME", "pod-2")
	if err := r.Register(svc2); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}

	// check pod-1 has correct labels/annotations
	p1 := mockClient.Pods["pod-1"]
	if svcLabel1, ok := p1.Metadata.Labels[svcSelectorPrefix+"foo.service"]; !ok || *svcLabel1 != svcSelectorValue {
		t.Fatalf("expected to have pod selector label for foo-service")
	}
	if _, ok := p1.Metadata.Labels[svcSelectorPrefix+"bar.service"]; ok {
		t.Fatal("pod 1 shouldnt have label for bar-service")
	}

	// check pod-2 has correct labels/annotations
	p2 := mockClient.Pods["pod-2"]
	if svcLabel2, ok := p2.Metadata.Labels[svcSelectorPrefix+"bar.service"]; !ok || *svcLabel2 != svcSelectorValue {
		t.Fatalf("expected to have pod selector label for bar-service")
	}
	if _, ok := p2.Metadata.Labels[svcSelectorPrefix+"foo.service"]; ok {
		t.Fatal("pod 2 shouldnt have label for foo-service")
	}

	// check k8s service one has expected result
	s1, ok := mockClient.Endpoints["foo-service"]
	if !ok {
		t.Fatalf("expected foo-service to exist")
	}
	svcData1, ok := s1.Metadata.Annotations[annotationServiceKeyPrefix+"pod-1"]
	if !ok || len(*svcData1) == 0 {
		t.Fatalf("expected to have annotation")
	}
	if _, okp2 := s1.Metadata.Annotations[annotationServiceKeyPrefix+"pod-2"]; okp2 {
		t.Fatal("foo-service shouldnt have annotation for pod2")
	}

	// check k8s service one has expected result
	s2, ok := mockClient.Endpoints["bar-service"]
	if !ok {
		t.Fatalf("expected bar-service to exist")
	}
	svcData2, ok := s2.Metadata.Annotations[annotationServiceKeyPrefix+"pod-2"]
	if !ok || len(*svcData2) == 0 {
		t.Fatalf("expected to have annotation")
	}
	if _, okp1 := s2.Metadata.Annotations[annotationServiceKeyPrefix+"pod-1"]; okp1 {
		t.Fatal("bar-service shouldnt have annotation for pod1")
	}

	// unmarshal service data from annotation and compare
	// service passed in to .Register()
	var service1 *registry.Service
	if err := json.Unmarshal([]byte(*svcData1), &service1); err != nil {
		t.Fatalf("did not expect register unmarshal to fail %v", err)
	}
	if !reflect.DeepEqual(svc1, service1) {
		t.Fatal("services did not match")
	}

	var service2 *registry.Service
	if err := json.Unmarshal([]byte(*svcData2), &service2); err != nil {
		t.Fatalf("did not expect register unmarshal to fail %v", err)
	}
	if !reflect.DeepEqual(svc2, service2) {
		t.Fatal("services did not match")
	}
}

func TestRegisterSingleVersionedServiceTwoPods(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	svc1 := testdata["foo.service"][0]
	svc2 := testdata["foo.service"][1]

	r := newMockRegistry()

	os.Setenv("HOSTNAME", "pod-1")
	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}
	os.Setenv("HOSTNAME", "pod-2")
	if err := r.Register(svc2); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}

	// check pod-1 has correct labels/annotations
	p1 := mockClient.Pods["pod-1"]
	if svcLabel1, ok := p1.Metadata.Labels[svcSelectorPrefix+"foo.service"]; !ok || *svcLabel1 != svcSelectorValue {
		t.Fatalf("expected to have pod selector label for foo-service")
	}

	// check pod-2 has correct labels/annotations
	p2 := mockClient.Pods["pod-2"]
	if svcLabel2, ok := p2.Metadata.Labels[svcSelectorPrefix+"foo.service"]; !ok || *svcLabel2 != svcSelectorValue {
		t.Fatalf("expected to have pod selector label for foo-service")
	}

	// check k8s service one has expected result
	s, ok := mockClient.Endpoints["foo-service"]
	if !ok {
		t.Fatalf("expected foo-service to exist")
	}
	svcData1, ok := s.Metadata.Annotations[annotationServiceKeyPrefix+"pod-1"]
	if !ok || len(*svcData1) == 0 {
		t.Fatalf("expected to have annotation")
	}
	svcData2, ok := s.Metadata.Annotations[annotationServiceKeyPrefix+"pod-2"]
	if !ok || len(*svcData2) == 0 {
		t.Fatalf("expected to have annotation")
	}

	// unmarshal service data from annotation and compare
	// service passed in to .Register()
	var service1 *registry.Service
	if err := json.Unmarshal([]byte(*svcData1), &service1); err != nil {
		t.Fatalf("did not expect register unmarshal to fail %v", err)
	}
	if !reflect.DeepEqual(svc1, service1) {
		t.Fatal("services did not match")
	}

	var service2 *registry.Service
	if err := json.Unmarshal([]byte(*svcData2), &service2); err != nil {
		t.Fatalf("did not expect register unmarshal to fail %v", err)
	}
	if !reflect.DeepEqual(svc2, service2) {
		t.Fatal("services did not match")
	}
}

func TestDeregister(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	os.Setenv("HOSTNAME", "pod-1")
	svc1 := testdata["foo.service"][0]
	svc2 := testdata["foo.service"][1]

	r := newMockRegistry()

	os.Setenv("HOSTNAME", "pod-1")
	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}
	os.Setenv("HOSTNAME", "pod-2")
	if err := r.Register(svc2); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}

	// deregister one service
	os.Setenv("HOSTNAME", "pod-1")
	if err := r.Deregister(svc1); err != nil {
		t.Fatalf("did not expect Deregister to fail %v", err)
	}

	// check pod-1 has correct labels/annotations
	p1 := mockClient.Pods["pod-1"]
	if svcLabel1, ok := p1.Metadata.Labels[svcSelectorPrefix+"foo.service"]; ok && *svcLabel1 == svcSelectorValue {
		t.Fatalf("expected to NOT have pod selector label for foo-service")
	}

	// check pod-2 has correct labels/annotations
	p2 := mockClient.Pods["pod-2"]
	if svcLabel2, ok := p2.Metadata.Labels[svcSelectorPrefix+"foo.service"]; !ok || *svcLabel2 != svcSelectorValue {
		t.Fatalf("expected to have pod selector label for foo-service")
	}

	// check k8s service one has expected result
	s, ok := mockClient.Endpoints["foo-service"]
	if !ok {
		t.Fatalf("expected foo-service to exist")
	}
	svcData1, ok := s.Metadata.Annotations[annotationServiceKeyPrefix+"pod-1"]
	if ok && len(*svcData1) != 0 {
		t.Fatalf("expected to NOT have annotation")
	}
	svcData2, ok := s.Metadata.Annotations[annotationServiceKeyPrefix+"pod-2"]
	if !ok || len(*svcData2) == 0 {
		t.Fatalf("expected to have annotation")
	}

}

func TestGetService(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	os.Setenv("HOSTNAME", "pod-1")
	svc1 := testdata["foo.service"][0]

	r := newMockRegistry()

	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect Register to fail %v", err)
	}

	service, err := r.GetService("foo.service")
	if err != nil {
		t.Fatalf("did not expect GetService to fail %v", err)
	}

	// compare services
	if !hasServices(service, []*registry.Service{svc1}) {
		t.Fatal("expected service to match")
	}
}

func TestGetServiceSameServiceTwoPods(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	svc1 := &registry.Service{
		Name:    "foo.service",
		Version: "1",
		Nodes: []*registry.Node{
			{
				Id:      "foo-service-1",
				Address: "10.0.0.1",
				Port:    80,
				Metadata: map[string]string{
					"foo": "bar",
				},
			},
		},
		Endpoints: []*registry.Endpoint{{
			Name: "foo-service.ep1",
		}},
	}

	r := newMockRegistry()

	os.Setenv("HOSTNAME", "pod-1")
	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect Register to fail %v", err)
	}
	os.Setenv("HOSTNAME", "pod-2")
	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect Register to fail %v", err)
	}

	service, err := r.GetService("foo.service")
	if err != nil {
		t.Fatalf("did not expect GetService to fail %v", err)
	}

	if len(service) != 1 {
		t.Fatal("expected there to be only 1 service")
	}

	if len(service[0].Nodes) != 2 {
		t.Fatal("expected there to be 2 nodes")
	}
	if !hasNodes(service[0].Nodes, []*registry.Node{
		&registry.Node{
			Id:      "foo-service-1",
			Address: "10.0.0.1",
			Port:    80,
			Metadata: map[string]string{
				"foo": "bar",
			},
		},
		&registry.Node{
			Id:      "foo-service-1",
			Address: "10.0.0.2",
			Port:    80,
			Metadata: map[string]string{
				"foo": "bar",
			},
		},
	}) {
		t.Fatal("nodes dont match")
	}
}

func TestGetServiceTwoVersionsTwoPods(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	svc1 := &registry.Service{
		Name:    "foo.service",
		Version: "1",
		Nodes: []*registry.Node{
			{
				Id:      "foo-service-1",
				Address: "10.0.0.1",
				Port:    80,
				Metadata: map[string]string{
					"v": "1",
				},
			},
		},
		Endpoints: []*registry.Endpoint{{
			Name: "foo-service.ep1",
		}},
	}
	svc2 := &registry.Service{
		Name:    "foo.service",
		Version: "2",
		Nodes: []*registry.Node{
			{
				Id:      "foo-service-1",
				Address: "10.0.0.2",
				Port:    80,
				Metadata: map[string]string{
					"v": "2",
				},
			},
		},
		Endpoints: []*registry.Endpoint{{
			Name: "foo-service.ep1",
		}, {
			Name: "foo-service.ep1",
		}},
	}

	r := newMockRegistry()

	os.Setenv("HOSTNAME", "pod-1")
	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect Register to fail %v", err)
	}
	os.Setenv("HOSTNAME", "pod-2")
	if err := r.Register(svc2); err != nil {
		t.Fatalf("did not expect Register to fail %v", err)
	}

	service, err := r.GetService("foo.service")
	if err != nil {
		t.Fatalf("did not expect GetService to fail %v", err)
	}

	if len(service) != 2 {
		t.Fatal("expected there to be 2 services")
	}

	// compare services
	if !hasServices(service, []*registry.Service{svc1, svc2}) {
		t.Fatal("expected service to match")
	}
}

func TestListServices(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	svc1 := testdata["foo.service"][0]
	svc2 := testdata["bar.service"][0]

	r := newMockRegistry()

	os.Setenv("HOSTNAME", "pod-1")
	if err := r.Register(svc1); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}
	os.Setenv("HOSTNAME", "pod-2")
	if err := r.Register(svc2); err != nil {
		t.Fatalf("did not expect register to fail %v", err)
	}

	services, err := r.ListServices()
	if err != nil {
		t.Fatalf("did not expect ListServices to fail %v", err)
	}
	if !hasServices(services, []*registry.Service{
		{Name: "foo.service"},
		{Name: "bar.service"},
	}) {
		t.Fatal("expected services to equal")
	}

	os.Setenv("HOSTNAME", "pod-1")
	r.Deregister(svc1)
	services2, err := r.ListServices()
	if err != nil {
		t.Fatalf("did not expect ListServices to fail %v", err)
	}
	if !hasServices(services2, []*registry.Service{
		{Name: "bar.service"},
	}) {
		t.Fatal("expected services to equal")
	}

	// kill pod without deregistering.
	delete(mockClient.Pods, "pod-2")

	// shoudnt return old data
	services3, err := r.ListServices()
	if err != nil {
		t.Fatalf("did not expect ListServices to fail %v", err)
	}
	if len(services3) != 0 {
		t.Fatal("expected there to be no services")
	}

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

func TestWatcher(t *testing.T) {
	mockK8s()
	defer teardownK8s()
	os.Setenv("HOSTNAME", "pod-1")

	r := newMockRegistry()
	c := cache.NewSelector(selector.Registry(r))
	defer c.Close()

	time.Sleep(time.Millisecond)

	svc1 := testdata["foo.service"][0]
	svc2 := testdata["foo.service"][1]

	os.Setenv("HOSTNAME", "pod-1")
	r.Register(svc1)

	var wg sync.WaitGroup
	wg.Add(3)

	c.Select("foo.service", selector.WithFilter(func(svcs []*registry.Service) []*registry.Service {
		defer wg.Done()
		if !hasServices(svcs, []*registry.Service{svc1}) {
			t.Fatal("expected services to match")
		}
		return nil
	}))

	os.Setenv("HOSTNAME", "pod-2")
	r.Register(svc2)
	eps2, _ := mockClient.GetEndpoints("foo-service")
	mockClient.WatchResults <- watch.Event{
		Type:   watch.Modified,
		Object: *eps2,
	}

	// sleep to allow event to catchup
	time.Sleep(time.Millisecond)

	c.Select("foo.service", selector.WithFilter(func(svcs []*registry.Service) []*registry.Service {
		defer wg.Done()
		if !hasNodes(svcs[0].Nodes, []*registry.Node{svc1.Nodes[0], svc2.Nodes[0]}) {
			t.Fatal("expected to have same nodes")
		}
		return nil
	}))

	os.Setenv("HOSTNAME", "pod-1")
	r.Deregister(svc1)
	eps3, _ := mockClient.GetEndpoints("foo-service")
	mockClient.WatchResults <- watch.Event{
		Type:   watch.Modified,
		Object: *eps3,
	}

	// sleep to allow event to catchup
	time.Sleep(time.Millisecond)

	c.Select("foo.service", selector.WithFilter(func(svcs []*registry.Service) []*registry.Service {
		defer wg.Done()
		if !hasServices(svcs, []*registry.Service{svc2}) {
			t.Fatal("expected services to match")
		}
		return nil
	}))

	teardownK8s()
	mockClient.WatchResults <- watch.Event{
		Type: watch.Deleted,
		Object: client.Endpoints{
			Metadata: &client.Meta{
				Labels: map[string]*string{
					labelNameKey: &svc1.Name,
				},
			},
		},
	}

	// sleep to allow event to catchup
	time.Sleep(time.Millisecond)

	_, err := c.Select("foo.service")
	if err != registry.ErrNotFound {
		log.Fatal("expected registry.ErrNotFound")
	}

	out := make(chan bool)
	go func() {
		wg.Wait()
		close(out)
	}()

	select {
	case <-out:
		return
	case <-time.After(time.Second):
		t.Fatal("expected c.Select() to be called 3 times")
	}

}
