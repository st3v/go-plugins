package kubernetes

import (
	"errors"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/registry/kubernetes/client"
)

type k8sWatcher struct {
	registry *kregistry
	wr       *client.WatchRequest
	next     chan *registry.Result
}

// update decomposes the event from the k8s client
// and builds a registry.Result from it.
func (k *k8sWatcher) update(event *client.Event) {
	var action string
	switch event.Type {
	case "ADDED":
		action = "create"
	case "MODIFIED":
		action = "update"
	case "DELETED":
		action = "delete"
	default:
		return
	}

	ks := &registry.Service{
		Name: event.Object.Metadata.Name,
	}

	for _, item := range event.Object.Subsets {
		p := 80

		if len(item.Ports) > 0 {
			p = item.Ports[0].Port
		}

		for _, address := range item.Addresses {
			ks.Nodes = append(ks.Nodes, &registry.Node{
				Address: address.IP,
				Port:    p,
			})
		}
	}

	k.next <- &registry.Result{
		Action:  action,
		Service: ks,
	}
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
	close(k.next)
}

func newWatcher(kr *kregistry) (registry.Watcher, error) {
	// Create watch request
	wr, err := kr.client.WatchEndpoints()
	if err != nil {
		return nil, err
	}

	w := &k8sWatcher{
		registry: kr,
		wr:       wr,
		next:     make(chan *registry.Result, 10),
	}

	// range over watch request changes, and invoke
	// the update event
	go func() {
		for event := range wr.Change() {
			w.update(event)
		}
	}()

	return w, nil
}
