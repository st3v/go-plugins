package etcdv3

import (
	"errors"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/micro/go-micro/registry"
	"golang.org/x/net/context"
)

type etcdv3Watcher struct {
	once sync.Once
	stop chan bool
	w    clientv3.WatchChan
}

func newEtcdv3Watcher(r *etcdv3Registry) (registry.Watcher, error) {
	var once sync.Once
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan bool, 1)

	go func() {
		<-stop
		cancel()
	}()

	return &etcdv3Watcher{
		once: once,
		stop: stop,
		w:    r.client.Watch(ctx, prefix, clientv3.WithPrefix()),
	}, nil
}

func (ew *etcdv3Watcher) Next() (*registry.Result, error) {
	for wresp := range ew.w {
		if wresp.Err() != nil {
			return nil, wresp.Err()
		}
		for _, ev := range wresp.Events {
			service := decode(ev.Kv.Value)
			switch ev.Type {
			case clientv3.EventTypePut, clientv3.EventTypeDelete:
				var action string

				if ev.Type == clientv3.EventTypePut {
					if ev.IsCreate() {
						action = "create"
					} else if ev.IsModify() {
						action = "update"
					}
				}
				if ev.Type == clientv3.EventTypeDelete {
					action = "delete"
				}

				if service == nil {
					continue
				}

				return &registry.Result{
					Action:  action,
					Service: service,
				}, nil
			}
		}
	}
	return nil, errors.New("could not get next")
}

func (ew *etcdv3Watcher) Stop() {
	ew.once.Do(func() {
		ew.stop <- true
	})
}
