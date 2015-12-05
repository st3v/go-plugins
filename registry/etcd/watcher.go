package etcd

import (
	"errors"
	"sync"

	etcd "github.com/coreos/etcd/client"
	"github.com/micro/go-micro/registry"
	"golang.org/x/net/context"
)

type etcdWatcher struct {
	ctx  context.Context
	once sync.Once
	stop chan bool
	w    etcd.Watcher
}

func newEtcdWatcher(r *etcdRegistry) (registry.Watcher, error) {
	var once sync.Once
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan bool, 1)

	go func() {
		<-stop
		cancel()
	}()

	return &etcdWatcher{
		ctx:  ctx,
		w:    r.client.Watcher(prefix, &etcd.WatcherOptions{AfterIndex: 0, Recursive: true}),
		once: once,
		stop: stop,
	}, nil
}

func (ew *etcdWatcher) Next() (*registry.Result, error) {
	for {
		rsp, err := ew.w.Next(ew.ctx)
		if err != nil && ew.ctx.Err() != nil {
			return nil, err
		}

		if rsp.Node.Dir {
			continue
		}

		service := decode(rsp.Node.Value)
		if service == nil {
			switch {
			case rsp.Action != "delete":
				continue
			case rsp.PrevNode == nil:
				continue
			}
			// last ditch effort
			service = decode(rsp.PrevNode.Value)
			if service == nil {
				continue
			}
		}

		switch rsp.Action {
		case "set", "delete", "create", "update":
			if rsp.Action == "set" {
				rsp.Action = "update"
			}
			return &registry.Result{
				Action:  rsp.Action,
				Service: service,
			}, nil
		default:
			continue
		}

	}
	return nil, errors.New("could not get next")
}

func (ew *etcdWatcher) Stop() {
	ew.once.Do(func() {
		ew.stop <- true
	})
}
