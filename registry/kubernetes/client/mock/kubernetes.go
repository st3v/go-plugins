package mock

import (
	"sync"

	"github.com/micro/go-plugins/registry/kubernetes/client"
	"github.com/micro/go-plugins/registry/kubernetes/client/api"
	"github.com/micro/go-plugins/registry/kubernetes/client/watch"
)

// Client ...
type Client struct {
	sync.Mutex
	Pods         map[string]*client.Pod
	Services     map[string]*mockService
	Endpoints    map[string]*mockEndpoints
	WatchResults chan watch.Event
}

type mockMeta struct {
	Name        string
	Annotations map[string]*string
	Labels      *map[string]*string
}

type mockService struct {
	Metadata *mockMeta
	Spec     *client.ServiceSpec
}

type mockEndpoints struct {
	Metadata *mockMeta
	Subsets  []client.Subset
}

// UpdatePod ...
func (m *Client) UpdatePod(podName string, pod *client.Pod) (*client.Pod, error) {
	p, ok := m.Pods[podName]
	if !ok {
		return nil, api.ErrNotFound
	}

	updateMetadata(p.Metadata, pod.Metadata)
	return nil, nil
}

// GetService ...
func (m *Client) GetService(name string) (*client.Service, error) {

	s, ok := m.Services[name]
	if !ok {
		return nil, api.ErrNotFound
	}

	ss := &client.Service{
		Metadata: &client.Meta{
			Name:        s.Metadata.Name,
			Annotations: s.Metadata.Annotations,
			Labels:      *s.Metadata.Labels,
		},
		Spec: s.Spec,
	}

	return ss, nil
}

// CreateService ...
func (m *Client) CreateService(s *client.Service) (*client.Service, error) {

	m.Services[s.Metadata.Name] = &mockService{
		Metadata: &mockMeta{
			Name:        s.Metadata.Name,
			Annotations: s.Metadata.Annotations,
			Labels:      &s.Metadata.Labels,
		},
		Spec: s.Spec,
	}

	m.Endpoints[s.Metadata.Name] = &mockEndpoints{
		Metadata: &mockMeta{
			Name:        s.Metadata.Name,
			Annotations: s.Metadata.Annotations,
			Labels:      &s.Metadata.Labels,
		},
	}
	return s, nil
}

// UpdateService ...
func (m *Client) UpdateService(name string, s *client.Service) (*client.Service, error) {

	svc := m.Services[name]
	if svc == nil {
		return nil, api.ErrNotFound
	}

	updateMockMetadata(svc.Metadata, s.Metadata)
	return m.GetService(name)
}

// GetEndpoints ...
func (m *Client) GetEndpoints(name string) (*client.Endpoints, error) {

	svc, ok := m.Services[name]
	if !ok {
		return nil, api.ErrNotFound
	}

	selectors := svc.Spec.Selector
	sub := client.Subset{}

	for _, p := range m.Pods {
		if !labelFilterMatch(p.Metadata.Labels, selectors) {
			continue
		}

		// add address to subset
		sub.Addresses = append(sub.Addresses, client.Address{
			IP:        p.Status.PodIP,
			TargetRef: client.ObjectRef{Name: p.Metadata.Name},
		})
	}

	// sub.Ports = []client.SubsetPort{}
	for _, p := range svc.Spec.Ports {
		sub.Ports = append(sub.Ports, client.SubsetPort{
			Name: p.Name,
			Port: p.Port,
		})
	}

	eps, ok := m.Endpoints[name]
	if !ok {
		return nil, api.ErrNotFound
	}

	e := &client.Endpoints{
		Metadata: &client.Meta{
			Name:        eps.Metadata.Name,
			Annotations: eps.Metadata.Annotations,
			Labels:      *eps.Metadata.Labels,
		},
		Subsets: []client.Subset{sub},
	}
	return e, nil
}

// UpdateEndpoints ...
func (m *Client) UpdateEndpoints(name string, s *client.Endpoints) (*client.Endpoints, error) {
	eps := m.Endpoints[name]
	if eps == nil {
		return nil, api.ErrNotFound
	}

	updateMockMetadata(eps.Metadata, s.Metadata)
	return m.GetEndpoints(name)
}

// ListEndpoints ...
func (m *Client) ListEndpoints(labels map[string]string) (*client.EndpointsList, error) {
	var svcs []client.Endpoints
	for _, v := range m.Endpoints {
		if labelFilterMatch(*v.Metadata.Labels, labels) {
			e, err := m.GetEndpoints(v.Metadata.Name)
			if err != nil {
				return nil, err
			}
			svcs = append(svcs, *e)
		}
	}
	return &client.EndpointsList{
		Items: svcs,
	}, nil
}

// WatchEndpoints ...
func (m *Client) WatchEndpoints(labels map[string]string) (watch.Watch, error) {
	w := &mockWatcher{
		results: make(chan watch.Event),
		stop:    make(chan bool),
	}

	go func() {
		for {
			select {
			case f := <-m.WatchResults:
				w.results <- f
			case <-w.stop:
				break
			}
		}
	}()

	return w, nil
}

// newClient ...
func newClient() client.Kubernetes {
	return &Client{}
}

// NewClient ...
func NewClient() *Client {
	return &Client{
		Pods:         make(map[string]*client.Pod),
		Services:     make(map[string]*mockService),
		Endpoints:    make(map[string]*mockEndpoints),
		WatchResults: make(chan watch.Event),
	}
}

// Teardown ...
func Teardown(c *Client) {
	c.Pods = make(map[string]*client.Pod)
	c.Services = make(map[string]*mockService)
	c.Endpoints = make(map[string]*mockEndpoints)
}
