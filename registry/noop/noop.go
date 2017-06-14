// Package noop is a registry which does nothing
package noop

import (
	"github.com/micro/go-micro/registry"
)

type noopRegistry struct{}

func (m *noopRegistry) GetService(service string) ([]*registry.Service, error) {
	return nil, nil
}

func (m *noopRegistry) ListServices() ([]*registry.Service, error) {
	return nil, nil
}

func (m *noopRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	return nil
}

func (m *noopRegistry) Deregister(s *registry.Service) error {
	return nil
}

func (m *noopRegistry) Watch() (registry.Watcher, error) {
	return &noopWatcher{exit: make(chan bool)}, nil
}

func (m *noopRegistry) String() string {
	return "noop"
}

func NewRegistry() registry.Registry {
	return &noopRegistry{}
}
