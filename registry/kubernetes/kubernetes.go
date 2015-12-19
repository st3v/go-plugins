package kubernetes

import (
	"fmt"
	"os"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"

	api "k8s.io/kubernetes/pkg/api/unversioned"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

type kregistry struct {
	client    *k8s.Client
	namespace string
}

func init() {
	cmd.Registries["kubernetes"] = NewRegistry
}

func (c *kregistry) Deregister(s *registry.Service) error {
	return nil
}

func (c *kregistry) Register(s *registry.Service) error {
	return nil
}

func (c *kregistry) GetService(name string) ([]*registry.Service, error) {
	selector := labels.SelectorFromSet(labels.Set{"name": name})
	lb := api.LabelSelector{selector}
	fd := api.FieldSelector{fields.Everything()}
	services, err := c.client.Services(c.namespace).List(api.ListOptions{LabelSelector: lb, FieldSelector: fd})
	if err != nil {
		return nil, err
	}

	if len(services.Items) == 0 {
		return nil, fmt.Errorf("Service not found")
	}

	ks := &registry.Service{
		Name: name,
	}

	for _, item := range services.Items {
		ks.Nodes = append(ks.Nodes, &registry.Node{
			Address: item.Spec.ClusterIP,
			Port:    item.Spec.Ports[0].Port,
		})
	}

	return []*registry.Service{ks}, nil
}

func (c *kregistry) ListServices() ([]*registry.Service, error) {
	var services []*registry.Service
	lb := api.LabelSelector{labels.Everything()}
	fd := api.FieldSelector{fields.Everything()}

	rsp, err := c.client.Services(c.namespace).List(api.ListOptions{LabelSelector: lb, FieldSelector: fd})
	if err != nil {
		return nil, err
	}

	for _, svc := range rsp.Items {
		if len(svc.ObjectMeta.Labels["name"]) == 0 {
			continue
		}

		services = append(services, &registry.Service{
			Name: svc.ObjectMeta.Labels["name"],
		})
	}

	return services, nil
}

func (c *kregistry) Watch() (registry.Watcher, error) {
	return newWatcher(c)
}

func (c *kregistry) String() string {
	return "kubernetes"
}

func NewRegistry(addrs []string, opts ...registry.Option) registry.Registry {
	host := "http://" + os.Getenv("KUBERNETES_RO_SERVICE_HOST") + ":" + os.Getenv("KUBERNETES_RO_SERVICE_PORT")
	if len(addrs) > 0 {
		host = addrs[0]
	}

	client, _ := k8s.New(&k8s.Config{
		Host: host,
	})

	kr := &kregistry{
		client:    client,
		namespace: "default",
	}

	return kr
}
