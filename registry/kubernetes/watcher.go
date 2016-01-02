package kubernetes

import (
	"errors"
	"net"

	"github.com/micro/go-micro/registry"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type watcher struct {
	registry *kregistry
	watcher  watch.Interface
	next     chan *registry.Result
}

func (k *watcher) update(event watch.Event) {
	if event.Object == nil {
		return
	}

	var service *api.Service
	switch obj := event.Object.(type) {
	case *api.Service:
		service = obj
	default:
		return
	}

	name, exists := service.ObjectMeta.Labels["name"]
	if !exists {
		return
	}

	var action string
	switch event.Type {
	case watch.Added:
		action = "create"
	case watch.Modified:
		action = "update"
	case watch.Deleted:
		action = "delete"
	default:
		return
	}

	serviceIP := net.ParseIP(service.Spec.ClusterIP)

	k.next <- &registry.Result{
		Action: action,
		Service: &registry.Service{
			Name: name,
			Nodes: []*registry.Node{
				&registry.Node{
					Address: serviceIP.String(),
					Port:    service.Spec.Ports[0].Port,
				},
			},
		},
	}
}

func (k *watcher) Next() (*registry.Result, error) {
	r, ok := <-k.next
	if !ok {
		return nil, errors.New("result chan closed")
	}
	return r, nil
}

func (k *watcher) Stop() {
	k.watcher.Stop()
	close(k.next)
}

func newWatcher(kr *kregistry) (registry.Watcher, error) {
	svi := kr.client.Services(api.NamespaceAll)

	lb := labels.Everything()
	fd := fields.Everything()
	services, err := svi.List(api.ListOptions{LabelSelector: lb, FieldSelector: fd})
	if err != nil {
		return nil, err
	}

	watch, err := svi.Watch(api.ListOptions{LabelSelector: lb, FieldSelector: fd, ResourceVersion: services.ResourceVersion})
	if err != nil {
		return nil, err
	}

	w := &watcher{
		registry: kr,
		watcher:  watch,
		next:     make(chan *registry.Result, 10),
	}

	go func() {
		for event := range watch.ResultChan() {
			w.update(event)
		}
	}()

	return w, nil
}
