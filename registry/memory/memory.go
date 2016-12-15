package memory

import (
	"sync"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
	"golang.org/x/net/context"
)

type memoryRegistry struct {
	sync.RWMutex
	services map[string][]*registry.Service
}

func init() {
	cmd.DefaultRegistries["memory"] = NewRegistry
}

func (m *memoryRegistry) GetService(service string) ([]*registry.Service, error) {
	m.RLock()
	defer m.RUnlock()

	s, ok := m.services[service]
	if !ok || len(s) == 0 {
		return nil, registry.ErrNotFound
	}
	return s, nil

}

func (m *memoryRegistry) ListServices() ([]*registry.Service, error) {
	m.RLock()
	defer m.RUnlock()

	var services []*registry.Service
	for _, service := range m.services {
		services = append(services, service...)
	}
	return services, nil
}

func (m *memoryRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	m.Lock()
	defer m.Unlock()

	services := addServices(m.services[s.Name], []*registry.Service{s})
	m.services[s.Name] = services
	return nil
}

func (m *memoryRegistry) Deregister(s *registry.Service) error {
	m.Lock()
	defer m.Unlock()

	services := delServices(m.services[s.Name], []*registry.Service{s})
	m.services[s.Name] = services
	return nil
}

func (m *memoryRegistry) Watch() (registry.Watcher, error) {
	return &memoryWatcher{exit: make(chan bool)}, nil
}

func (m *memoryRegistry) String() string {
	return "memory"
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	options := registry.Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	services := getServices(options.Context)
	if services == nil {
		services = make(map[string][]*registry.Service)
	}

	return &memoryRegistry{
		services: services,
	}
}
