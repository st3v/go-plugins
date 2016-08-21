package eureka

import (
	"errors"
	"testing"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/eureka/mock"
)

func TestRegistration(t *testing.T) {
	testData := []struct {
		appInstanceErr       error
		callCountAppInstance int
		callCountRegister    int
		callCountHeartbeat   int
	}{
		{errors.New("Instance not existing"), 1, 1, 0}, // initial register
		{nil, 1, 0, 1},                                 // subsequent register
	}

	eureka := NewRegistry().(*eurekaRegistry)

	service := &registry.Service{
		Nodes: []*registry.Node{new(registry.Node)},
	}

	for _, test := range testData {
		mockClient := new(mock.Client)
		mockClient.AppInstanceReturns(nil, test.appInstanceErr)
		eureka.client = mockClient

		eureka.Register(service)

		if mockClient.AppInstanceCallCount() != test.callCountAppInstance {
			t.Errorf(
				"Expected exactly %d calls to AppInstance, got %d calls.",
				test.callCountAppInstance,
				mockClient.AppInstanceCallCount(),
			)
		}

		if mockClient.RegisterCallCount() != test.callCountRegister {
			t.Errorf(
				"Expected exactly %d calls of RegisterInstance, got %d calls.",
				test.callCountRegister,
				mockClient.RegisterCallCount(),
			)
		}

		if mockClient.HeartbeatCallCount() != test.callCountHeartbeat {
			t.Errorf(
				"Expected exactly %d calls of Heartbeat, got %d calls.",
				test.callCountHeartbeat,
				mockClient.HeartbeatCallCount(),
			)
		}
	}
}
