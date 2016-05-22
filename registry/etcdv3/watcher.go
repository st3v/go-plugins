package etcdv3

import (
	"errors"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/micro/go-micro/registry"
	"golang.org/x/net/context"
)

type etcdv3Watcher struct {
	stop    chan bool
	w       clientv3.WatchChan
	client  *clientv3.Client
	timeout time.Duration
}

func newEtcdv3Watcher(r *etcdv3Registry, timeout time.Duration) (registry.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan bool, 1)

	go func() {
		<-stop
		cancel()
	}()

	return &etcdv3Watcher{
		stop:    stop,
		w:       r.client.Watch(ctx, prefix, clientv3.WithPrefix()),
		client:  r.client,
		timeout: timeout,
	}, nil
}

func (ew *etcdv3Watcher) Next() (*registry.Result, error) {
	for wresp := range ew.w {
		if wresp.Err() != nil {
			return nil, wresp.Err()
		}
		for _, ev := range wresp.Events {
			service := decode(ev.Kv.Value)
			var action string

			switch ev.Type {
			case clientv3.EventTypePut:
				if ev.IsCreate() {
					action = "create"
				} else if ev.IsModify() {
					action = "update"
				}
			case clientv3.EventTypeDelete:
				action = "delete"

				// get the cached value
				ctx, cancel := context.WithTimeout(context.Background(), ew.timeout)
				defer cancel()

				resp, err := ew.client.Get(ctx, path.Join(cachePrefix, string(ev.Kv.Key)))
				if err != nil {
					return nil, err
				}

				for _, ev := range resp.Kvs {
					service = decode(ev.Value)
				}

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
	return nil, errors.New("could not get next")
}

func (ew *etcdv3Watcher) Stop() {
	select {
	case <-ew.stop:
		return
	default:
		close(ew.stop)
	}
}
