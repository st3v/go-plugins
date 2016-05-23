package kubernetes

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/kubernetes/client"
	"github.com/micro/go-plugins/registry/kubernetes/client/watch"
)

type k8sWatcher struct {
	registry *kregistry
	wr       watch.Watch
	next     chan *registry.Result

	sync.RWMutex
	services map[string][]*registry.Service
}

// handleEvent will handle any event coming from the k8s api
// and will compare against a cached version so that it can send
// an individual result for each service node.
func (k *k8sWatcher) handleEvent(event *watch.Event) {
	endpoints, ok := event.Object.(client.Endpoints)
	if !ok {
		return
	}

	name, ok := endpoints.Metadata.Labels[labelNameKey]
	if !ok {
		return
	}

	switch event.Type {
	case watch.Added, watch.Modified:

	case watch.Deleted:
		// service was deleted, therefore all the nodes were too.
		// send in a blank
		k.next <- &registry.Result{
			Action:  "delete",
			Service: &registry.Service{Name: *name},
		}
		return
	default:
		return
	}

	// build services from endpoints data, not versioned.
	// each service should have one node only.
	var svcs []*registry.Service
	subset := endpoints.Subsets[0]

	if len(endpoints.Subsets) > 0 && len(subset.Addresses) > 0 && len(subset.Ports) > 0 {
		port := subset.Ports[0].Port

		// loop through endpoints
		for _, address := range subset.Addresses {
			podKey := annotationServiceKeyPrefix + address.TargetRef.Name
			svcStr, podKeyOk := endpoints.Metadata.Annotations[podKey]
			if !podKeyOk {
				continue
			}

			// TODO: only do this if the service node has changed
			// from the cached version. Maybe have a SHA annotation.
			var svc registry.Service
			err := json.Unmarshal([]byte(*svcStr), &svc)
			if err != nil {
				return
			}

			svc.Nodes = []*registry.Node{
				&registry.Node{
					Address:  address.IP,
					Port:     port,
					Metadata: svc.Nodes[0].Metadata,
					Id:       svc.Nodes[0].Id,
				},
			}

			svcs = append(svcs, &svc)
		}
	}

	// check cache
	k.Lock()
	cache, ok := k.services[*name]
	k.Unlock()

	if !ok {
		// service doesnt yet exist,
		// emit create for each service node.
		for _, svc := range svcs {
			k.next <- &registry.Result{
				Action:  "create",
				Service: svc,
			}
		}

		k.RLock()
		k.services[*name] = svcs
		k.RUnlock()
		return
	}

	// cache exists, trigger create/update/delete accordingly

	// find service in cache and emit update, otherwise emit create
	for _, svc := range svcs {
		found := false

		for _, cSvc := range cache {
			if cSvc.Nodes[0].Id == svc.Nodes[0].Id {
				found = true
				break
			}
		}

		if !found {
			// not in cache, so emit create
			k.next <- &registry.Result{
				Action:  "create",
				Service: svc,
			}
			continue
		}

		// in cache, so just send update
		// TODO: improve this, could do with some kind of hash
		// to compare against, and only send if changed
		k.next <- &registry.Result{
			Action:  "update",
			Service: svc,
		}
	}

	// remove items from cache no longer existing
	for _, cSvc := range cache {
		found := false

		for _, svc := range svcs {
			if cSvc.Nodes[0].Id == svc.Nodes[0].Id {
				found = true
				break
			}
		}

		if !found {
			k.next <- &registry.Result{
				Action:  "delete",
				Service: cSvc,
			}
		}
	}

	k.RLock()
	k.services[*name] = svcs
	k.RUnlock()

}

// Next will block until a new result comes in
func (k *k8sWatcher) Next() (*registry.Result, error) {
	r, ok := <-k.next
	if !ok {
		return nil, errors.New("result chan closed")
	}
	return r, nil
}

// Stop will cancel any requests, and close channels
func (k *k8sWatcher) Stop() {
	k.wr.Stop()

	select {
	case <-k.next:
		return
	default:
		close(k.next)
	}
}

func newWatcher(kr *kregistry) (registry.Watcher, error) {
	// Create watch request
	wr, err := kr.client.WatchEndpoints(map[string]string{
		labelTypeKey: labelTypeValueService,
	})
	if err != nil {
		return nil, err
	}

	k := &k8sWatcher{
		registry: kr,
		wr:       wr,
		next:     make(chan *registry.Result),
		services: make(map[string][]*registry.Service),
	}

	// range over watch request changes, and invoke
	// the update event
	go func() {
		for event := range wr.ResultChan() {
			k.handleEvent(&event)
		}
		// request was canceled.
		k.Stop()
	}()

	return k, nil
}
