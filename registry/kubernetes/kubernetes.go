package kubernetes

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-plugins/registry/kubernetes/client"
	"github.com/micro/go-plugins/registry/kubernetes/client/api"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
)

type kregistry struct {
	sync.Mutex
	client  client.Kubernetes
	timeout time.Duration
}

var (
	// used on pods as labels & services to select
	// eg: svcSelectorPrefix+"svc.name"
	svcSelectorPrefix = "micro.mu/selector-"
	svcSelectorValue  = "service"

	labelNameKey          = "micro.mu/name"
	labelTypeKey          = "micro.mu/type"
	labelTypeValueService = "service"
	labelTypeValuePod     = "pod"

	// used on k8s services to scope a serialised
	// micro service by pod name
	annotationServiceKeyPrefix = "micro.mu/service-"
)

func init() {
	cmd.DefaultRegistries["kubernetes"] = NewRegistry
}

var (
	re = regexp.MustCompile("[^-a-z0-9]+")
)

// converts names to RFC952 names by
//   - replacing '.' with '-'
//   - lowercasing
//   - stripping anything other [-a-z0-9]
func toSlug(s string) string {
	n := strings.Replace(s, ".", "-", -1)
	return re.ReplaceAllString(strings.ToLower(n), "")
}

// Register will
//   * get service endpoints if it exists, for use when cleaning up annotations
//   * create a new K8s service, unless it already exists
//   * updates service endpoints annotations with a serialised version
//     of the service being registered using the curernt pod as a key.
//   * add a service selector label to the current pod.
func (c *kregistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Nodes) != 1 {
		return errors.New("Kubernetes registry will only register one node at a time")
	}

	c.Lock()
	defer c.Unlock()

	// TODO: grab podname from somewhere better than this.
	podName := os.Getenv("HOSTNAME")
	svcName := s.Name
	svcPort := s.Nodes[0].Port

	// encode micro service
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	svc := string(b)

	// GetEndpoints
	eps, err := c.client.GetEndpoints(toSlug(svcName))
	if err == api.ErrNotFound {
		err = nil
		eps = nil
	} else if err != nil {
		return err
	}

	// if endpoints does not exist, create service
	if eps == nil {
		_, err := c.createK8sService(svcName, svcPort)
		if err != nil {
			return err
		}
	}

	// annotations to include when endpoints is updated
	annotations := map[string]*string{
		annotationServiceKeyPrefix + podName: &svc,
	}

	// if endpoints does exist, nil out any old annotations
	if eps != nil {
		if len(eps.Subsets) > 0 {
			subset := eps.Subsets[0]

			for k := range eps.Metadata.Annotations {
				if !strings.HasPrefix(k, annotationServiceKeyPrefix) {
					continue
				}

				found := false
				for _, address := range subset.Addresses {
					n := address.TargetRef.Name
					p := strings.TrimPrefix(k, annotationServiceKeyPrefix)
					if p == n {
						found = true
						break
					}
				}

				if !found {
					annotations[k] = nil
				}
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	errs := make(chan error)

	// Update Endpoints with annotations
	go func() {
		defer wg.Done()
		_, err := c.client.UpdateEndpoints(toSlug(svcName), &client.Endpoints{
			Metadata: &client.Meta{Annotations: annotations},
		})
		if err != nil {
			errs <- err
		}
	}()

	// Update pod with service selector
	go func() {
		defer wg.Done()
		_, err := c.client.UpdatePod(podName, &client.Pod{
			Metadata: &client.Meta{
				Labels: map[string]*string{
					labelTypeKey:                &labelTypeValuePod,
					svcSelectorPrefix + svcName: &svcSelectorValue,
				},
			},
		})
		if err != nil {
			errs <- err
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case err := <-errs:
		return err
	case <-time.After(c.timeout):
		return errors.New("timeout")
	}
}

func (c *kregistry) createK8sService(svcName string, port int) (*client.Service, error) {
	ksNew := &client.Service{
		Metadata: &client.Meta{
			Name: toSlug(svcName),
			Labels: map[string]*string{
				labelNameKey: &svcName,
				labelTypeKey: &labelTypeValueService,
			},
		},
		Spec: &client.ServiceSpec{
			// ClusterIP: "None",
			Ports: []client.Port{
				client.Port{
					Name:       "http",
					Port:       port,
					TargetPort: port,
				},
			},
			Selector: map[string]string{
				svcSelectorPrefix + svcName: svcSelectorValue,
			},
		},
	}

	return c.client.CreateService(ksNew)
}

// Deregister will remove a annotations from a service, and remove
// service selector label from the current pod
func (c *kregistry) Deregister(s *registry.Service) error {
	if len(s.Nodes) != 1 {
		return errors.New("Kubernetes registry will only deregister one node at a time")
	}

	c.Lock()
	defer c.Unlock()

	// TODO: grab podname from somewhere better than this.
	podName := os.Getenv("HOSTNAME")
	svcName := s.Name

	var wg sync.WaitGroup
	wg.Add(2)
	errs := make(chan error)

	// Update Endpoints with annotations
	go func() {
		defer wg.Done()
		_, err := c.client.UpdateEndpoints(toSlug(svcName), &client.Endpoints{
			Metadata: &client.Meta{
				Annotations: map[string]*string{
					annotationServiceKeyPrefix + podName: nil,
				},
			},
		})
		if err != nil {
			errs <- err
		}
	}()

	// Update pod with service selector
	go func() {
		defer wg.Done()
		_, err := c.client.UpdatePod(podName, &client.Pod{
			Metadata: &client.Meta{
				Labels: map[string]*string{
					svcSelectorPrefix + svcName: nil,
				},
			},
		})
		if err != nil {
			errs <- err
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case err := <-errs:
		return err
	case <-time.After(c.timeout):
		return errors.New("timeout")
	}
}

// GetService uses the endpoints API to retrieve the current service.
// It deserialises the micro services from the annotations,
// and uses the subset addresses to build a list of registry.Services
func (c *kregistry) GetService(name string) ([]*registry.Service, error) {
	endpoints, err := c.client.GetEndpoints(toSlug(name))
	if err == api.ErrNotFound {
		return nil, registry.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if len(endpoints.Subsets) == 0 || len(endpoints.Subsets[0].Addresses) == 0 {
		return nil, registry.ErrNotFound
	}

	// svcs mapped by version
	svcs := make(map[string]*registry.Service)
	subset := endpoints.Subsets[0]
	port := subset.Ports[0].Port

	// loop through endpoints
	for _, address := range subset.Addresses {
		// get serialised service from annotation
		an := annotationServiceKeyPrefix + address.TargetRef.Name
		svcStr, ok := endpoints.Metadata.Annotations[an]
		if !ok {
			continue
		}

		// unmarshal service string
		var svc registry.Service
		err := json.Unmarshal([]byte(*svcStr), &svc)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal service '%s' from service annotation '%s'", name, an)
		}

		// merge up pod service & ip with versioned service.
		vs, ok := svcs[svc.Version]
		if !ok {
			svc.Nodes[0].Address = address.IP
			svc.Nodes[0].Port = port
			svcs[svc.Version] = &svc
			continue
		}

		vs.Nodes = append(vs.Nodes, &registry.Node{
			Address:  address.IP,
			Port:     port,
			Metadata: svc.Nodes[0].Metadata,
			Id:       svc.Nodes[0].Id,
		})
	}

	var list []*registry.Service
	for _, val := range svcs {
		list = append(list, val)
	}
	return list, nil
}

// ListServices will list all the service names
func (c *kregistry) ListServices() ([]*registry.Service, error) {
	endpointsList, err := c.client.ListEndpoints(map[string]string{
		labelTypeKey: labelTypeValueService,
	})
	if err != nil {
		return nil, err
	}

	var list []*registry.Service

	for _, item := range endpointsList.Items {
		if len(item.Subsets) == 0 || len(item.Subsets[0].Addresses) == 0 {
			continue
		}
		name, ok := item.Metadata.Labels[labelNameKey]
		if !ok {
			continue
		}
		list = append(list, &registry.Service{Name: *name})
	}

	return list, nil
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
