package eureka

import (
	"errors"
	"time"

	"github.com/st3v/go-eureka"

	"github.com/micro/go-micro/registry"
)

type eurekaWatcher struct {
	watcher *eureka.Watcher
	exit    chan bool
}

func newWatcher(client eurekaClient) registry.Watcher {
	w := &eurekaWatcher{
		watcher: client.Watch(5 * time.Second),
		exit:    make(chan bool),
	}

	return w
}

func (e *eurekaWatcher) Stop() {
	e.watcher.Stop()
	close(e.exit)
}

func (e *eurekaWatcher) Next() (*registry.Result, error) {
	for {
		select {
		case <-e.exit:
			return nil, errors.New("watcher stopped")
		case event := <-e.watcher.Events():
			if r := result(event); r != nil {
				return r, nil
			}
		}
	}
}

func result(e eureka.Event) *registry.Result {
	var action string

	switch e.Type {
	case eureka.EventInstanceRegistered:
		action = "create"
	case eureka.EventInstanceDeregistered:
		action = "delete"
	default:
		action = "update"
	}

	service := appToService(&eureka.App{
		Name:      e.Instance.AppName,
		Instances: []*eureka.Instance{e.Instance},
	})

	if len(service) == 0 {
		return nil
	}

	return &registry.Result{
		Action:  action,
		Service: service[0],
	}
}
