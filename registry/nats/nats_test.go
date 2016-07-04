package nats_test

import (
	"testing"

	"code.posteo.de/go-common/test/assert"

	"github.com/micro/go-micro/registry"
)

func TestRegister(t *testing.T) {
	service := registry.Service{Name: "test"}
	assert.NoError(t, e.registryOne.Register(&service))
	defer e.registryOne.Deregister(&service)

	services, err := e.registryOne.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(services))

	services, err = e.registryTwo.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(services))
}

func TestDeregister(t *testing.T) {
	t.Skip("not properly implemented")

	service := registry.Service{Name: "test"}

	assert.NoError(t, e.registryOne.Register(&service))
	assert.NoError(t, e.registryOne.Deregister(&service))

	services, err := e.registryOne.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(services))

	services, err = e.registryTwo.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(services))
}

func TestGetService(t *testing.T) {
	services, err := e.registryTwo.GetService("one")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(services))
	assert.Equal(t, "one", services[0].Name)
	assert.Equal(t, 1, len(services[0].Nodes))
}

func TestGetServiceWithNoNodes(t *testing.T) {
	services, err := e.registryOne.GetService("missing")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(services))
}

func TestGetServiceFromMultipleNodes(t *testing.T) {
	services, err := e.registryOne.GetService("two")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(services))
	assert.Equal(t, "two", services[0].Name)
	assert.Equal(t, 2, len(services[0].Nodes))
}

func BenchmarkGetService(b *testing.B) {
	for n := 0; n < b.N; n++ {
		services, err := e.registryTwo.GetService("one")
		assert.NoError(b, err)
		assert.Equal(b, 1, len(services))
		assert.Equal(b, "one", services[0].Name)
	}
}

func BenchmarkGetServiceWithNoNodes(b *testing.B) {
	for n := 0; n < b.N; n++ {
		services, err := e.registryOne.GetService("missing")
		assert.NoError(b, err)
		assert.Equal(b, 0, len(services))
	}
}

func BenchmarkGetServiceFromMultipleNodes(b *testing.B) {
	for n := 0; n < b.N; n++ {
		services, err := e.registryTwo.GetService("two")
		assert.NoError(b, err)
		assert.Equal(b, 1, len(services))
		assert.Equal(b, "two", services[0].Name)
		assert.Equal(b, 2, len(services[0].Nodes))
	}
}
