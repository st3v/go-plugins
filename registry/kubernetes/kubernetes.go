package kubernetes

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-plugins/registry/kubernetes/client"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
)

type kregistry struct {
	client  client.Kubernetes
	timeout time.Duration
}

type podServiceMeta struct {
	Service  *registry.Service
	Metadata map[string]string
}

func init() {
	cmd.DefaultRegistries["kubernetes"] = NewRegistry
}

// Deregister shouldnt be required, as usually this will coincide
// with the pod and any annotations being destroyed along with it.
func (c *kregistry) Deregister(s *registry.Service) error {
	return nil
}

// Register sets annotations on the current pod
func (c *kregistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	//TODO: grab podname from somewhere better than this.
	podName := os.Getenv("HOSTNAME")

	eps, err := json.Marshal(s.Endpoints)
	if err != nil {
		return err
	}

	meta, err := json.Marshal(s.Nodes[0].Metadata)
	if err != nil {
		return err
	}

	pod := &client.Pod{
		Metadata: client.Meta{
			Annotations: map[string]string{
				"micro/name":      s.Name,
				"micro/version":   s.Version,
				"micro/meta":      strings.TrimSpace(string(meta)),
				"micro/endpoints": strings.TrimSpace(string(eps)),
			},
		},
	}

	return c.client.UpdatePod(podName, pod)
}

// GetService uses the `/endpoints/{name}` and `/pods?labelSelector=micro={name}`
// api to build the tree of Services, nodes, metadata and endpoints.
func (c *kregistry) GetService(name string) ([]*registry.Service, error) {
	errs := make(chan error, 1)
	endpoints := &client.Endpoints{}
	podList := &client.PodList{}

	var wg sync.WaitGroup
	wg.Add(2)

	// get all endpoints of pods, that match service name
	go func() {
		defer wg.Done()
		var err error
		endpoints, err = c.client.GetEndpoints(name)
		if err != nil {
			errs <- err
		}
	}()

	// fetch all pods with the label `micro=<service>`
	// TODO: make this label a setting?
	go func() {
		defer wg.Done()
		labels := map[string]string{"micro": name}
		var err error
		podList, err = c.client.GetPods(labels)
		if err != nil {
			errs <- err
		}
	}()

	done := make(chan bool, 2)
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(c.timeout):
		return nil, errors.New("GetService timed out")
	case err := <-errs:
		if err != nil {
			return nil, err
		}
	}

	if len(endpoints.Subsets) == 0 {
		return nil, errors.New("Service not found")
	}

	if len(podList.Items) == 0 {
		return nil, errors.New("No pods found with 'micro:<service>' label")
	}

	svcs := map[string]*registry.Service{}
	pods := map[string]*podServiceMeta{}

	// loop through all pods and add to map.
	for _, pod := range podList.Items {
		a := pod.Metadata.Annotations
		version := a["micro/version"]
		name := a["micro/name"]
		meta := a["micro/meta"]
		eps := a["micro/endpoints"]

		// get svc by version
		var svc = svcs[version]
		if svc == nil {

			endpoints := []*registry.Endpoint{}
			if len(eps) > 0 {
				err := json.Unmarshal([]byte(eps), &endpoints)
				if err != nil {
					return nil, err
				}
			}

			svc = &registry.Service{
				Name:      name,
				Version:   version,
				Endpoints: endpoints,
			}
			svcs[version] = svc
		}

		m := map[string]string{}
		if len(meta) > 0 {
			err := json.Unmarshal([]byte(meta), &m)
			if err != nil {
				return nil, err
			}
		}

		pods[pod.Metadata.Name] = &podServiceMeta{
			Service:  svc,
			Metadata: m,
		}

	}

	// loop through endpoints
	for _, item := range endpoints.Subsets {
		var p = 80

		if len(item.Ports) > 0 {
			p = item.Ports[0].Port
		}

		// each address has a targetRef.Name, which
		// references the pod name its assigned to.
		for _, address := range item.Addresses {
			// pick up the pod by ref from above
			pod := pods[address.TargetRef.Name]
			if pod == nil {
				continue
			}

			pod.Service.Nodes = append(pod.Service.Nodes, &registry.Node{
				Id:       address.TargetRef.Name,
				Address:  address.IP,
				Port:     p,
				Metadata: pod.Metadata,
			})
		}
	}

	var list []*registry.Service
	for _, val := range svcs {
		list = append(list, val)
	}
	return list, nil
}

// ListServices will list all the service names
func (c *kregistry) ListServices() ([]*registry.Service, error) {
	var services []*registry.Service
	var data *client.ServiceList
	errs := make(chan error, 1)

	go func() {
		var err error
		data, err = c.client.GetServices()
		errs <- err
	}()

	select {
	case <-time.After(c.timeout):
		return nil, errors.New("ListServices timed out")
	case err := <-errs:
		if err != nil {
			return nil, err
		}
	}

	for _, svc := range data.Items {
		services = append(services, &registry.Service{
			Name: svc.Metadata.Name,
		})
	}

	return services, nil
}

// Watch returns a kubernetes watcher
func (c *kregistry) Watch() (registry.Watcher, error) {
	return newWatcher(c)
}

func (c *kregistry) String() string {
	return "kubernetes"
}

// NewRegistry creates a kubernetes registry
func NewRegistry(opts ...registry.Option) registry.Registry {

	var options registry.Options
	for _, o := range opts {
		o(&options)
	}

	// get first host
	var host string
	if len(options.Addrs) > 0 && len(options.Addrs[0]) > 0 {
		host = options.Addrs[0]
	}

	if options.Timeout == 0 {
		options.Timeout = time.Second * 1
	}

	// if no hosts setup, assume InCluster
	var c client.Kubernetes
	if len(host) == 0 {
		c = client.NewClientInCluster()
	} else {
		c = client.NewClientByHost(host)
	}

	return &kregistry{
		client:  c,
		timeout: options.Timeout,
	}
}
