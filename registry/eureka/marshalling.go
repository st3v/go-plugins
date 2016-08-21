package eureka

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/st3v/go-eureka"

	"github.com/micro/go-micro/registry"
)

func appToService(app *eureka.App) []*registry.Service {
	serviceMap := make(map[string]*registry.Service)

	for _, instance := range app.Instances {
		var (
			id      = instance.ID
			addr    = instance.IPAddr
			port    = instance.Port
			version = instance.Metadata["version"]

			metadata  map[string]string
			endpoints []*registry.Endpoint
		)

		if k, ok := instance.Metadata["endpoints"]; ok {
			json.Unmarshal([]byte(k), &endpoints)
		}

		if k, ok := instance.Metadata["metadata"]; ok {
			json.Unmarshal([]byte(k), &metadata)
		}

		// get existing service
		service, ok := serviceMap[version]
		if !ok {
			// create new if doesn't exist
			service = &registry.Service{
				Name:      strings.ToLower(app.Name),
				Version:   version,
				Endpoints: endpoints,
			}
		}

		// append node
		service.Nodes = append(service.Nodes, &registry.Node{
			Id:       id,
			Address:  addr,
			Port:     int(port),
			Metadata: metadata,
		})

		// save
		serviceMap[version] = service
	}

	services := make([]*registry.Service, 0, len(serviceMap))
	for _, service := range serviceMap {
		services = append(services, service)
	}

	return services
}

// only parses first node
func serviceToInstance(service *registry.Service) (*eureka.Instance, error) {
	if len(service.Nodes) == 0 {
		return nil, errors.New("Require nodes")
	}

	node := service.Nodes[0]

	instance := &eureka.Instance{
		ID:             node.Id,
		AppName:        service.Name,
		HostName:       node.Address,
		IPAddr:         node.Address,
		VIPAddr:        node.Address,
		SecureVIPAddr:  node.Address,
		Port:           eureka.Port(node.Port),
		Status:         eureka.StatusUp,
		DataCenterInfo: eureka.DataCenter{Type: eureka.DataCenterTypePrivate},
		Metadata:       map[string]string{"version": service.Version},
	}

	// set endpoints
	if b, err := json.Marshal(service.Endpoints); err == nil {
		instance.Metadata["endpoints"] = string(b)
	}

	// set metadata
	if b, err := json.Marshal(node.Metadata); err == nil {
		instance.Metadata["metadata"] = string(b)
	}

	return instance, nil
}
