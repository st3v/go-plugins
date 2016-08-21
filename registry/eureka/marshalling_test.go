package eureka

import (
	"encoding/json"
	"testing"

	"github.com/micro/go-micro/registry"
	eureka "github.com/st3v/go-eureka"
)

func TestServiceToInstance(t *testing.T) {
	nodes := []*registry.Node{
		&registry.Node{
			Id:       "node0",
			Address:  "node0.example.com",
			Port:     1234,
			Metadata: map[string]string{"foo": "bar"},
		},
		&registry.Node{
			Id:      "node1",
			Address: "node1.example.com",
			Port:    9876,
		},
	}

	endpoints := []*registry.Endpoint{
		&registry.Endpoint{
			Name:     "endpoint",
			Request:  &registry.Value{"request-value", "request-value-type", []*registry.Value{}},
			Response: &registry.Value{"response-value", "response-value-type", []*registry.Value{}},
			Metadata: map[string]string{"endpoint-meta-key": "endpoint-meta-value"},
		},
	}

	service := &registry.Service{
		Name:      "service-name",
		Version:   "service-version",
		Nodes:     nodes,
		Endpoints: endpoints,
	}

	instance, err := serviceToInstance(service)
	if err != nil {
		t.Error("Unexpected serviceToInstance error:", err)
	}

	expectedEndpointsJSON, err := json.Marshal(endpoints)
	if err != nil {
		t.Error("Unexpected endpoints marshal error:", err)
	}

	expectedNodeMetadataJSON, err := json.Marshal(nodes[0].Metadata)
	if err != nil {
		t.Error("Unexpected node metadata marshal error:", err)
	}

	testData := []struct {
		name string
		want interface{}
		got  interface{}
	}{
		{"instance.AppName", service.Name, instance.AppName},
		{"instance.HostName", nodes[0].Address, instance.HostName},
		{"instance.IPAddr", nodes[0].Address, instance.IPAddr},
		{"instance.VipAddress", nodes[0].Address, instance.VIPAddr},
		{"instance.SecureVipAddress", nodes[0].Address, instance.SecureVIPAddr},
		{"instance.Port", nodes[0].Port, int(instance.Port)},
		{"instance.Status", eureka.StatusUp, instance.Status},
		{"instance.DataCenter.Name", eureka.DataCenterTypePrivate, instance.DataCenterInfo.Type},
		{`instance.Metadata["version"]`, service.Version, instance.Metadata["version"]},
		{`instance.Metadata["endpoints"]`, string(expectedEndpointsJSON), instance.Metadata["endpoints"]},
		{`instance.Metadata["metadata"]`, string(expectedNodeMetadataJSON), instance.Metadata["metadata"]},
	}

	for _, test := range testData {
		if test.got != test.want {
			t.Errorf("Unexpected %s: want %v, got %v", test.name, test.want, test.got)
		}
	}
}
