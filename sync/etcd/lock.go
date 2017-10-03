package etcd

import (
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/micro/go-os/sync"

	"golang.org/x/net/context"
)

type etcdLock struct {
	path string
	id   string
	node string

	api  client.KeysAPI
	opts sync.LockOptions
}

func (e *etcdLock) Id() string {
	return e.id
}

func (e *etcdLock) Acquire() error {
	opts := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       e.opts.TTL,
	}

	ctx, cancel := context.WithCancel(context.Background())

	if e.opts.Wait > 0 {
		go func() {
			<-time.After(e.opts.Wait)
			cancel()
		}()
	}

	// make path
	path := path.Join(e.path, strings.Replace(e.id, "/", "-", -1))

	for {
		_, err := e.api.Set(ctx, path, e.node, opts)
		if err == nil {
			return nil
		}

		// parse client error
		errr, ok := err.(client.Error)
		if !ok {
			return err
		}

		// if its not an already exist error return error
		if errr.Code != client.ErrorCodeNodeExist {
			return err
		}

		// get existing value
		rsp, err := e.api.Get(ctx, path, nil)
		if err != nil {
			return err
		}

		// create a watcher
		w := e.api.Watcher(path, &client.WatcherOptions{
			AfterIndex: rsp.Index,
			Recursive:  false,
		})

		// wait till key has been deleted
		for {
			rsp, err := w.Next(ctx)
			if err != nil {
				return err
			}

			if rsp.Action == "delete" || rsp.Action == "compareAndDelete" || rsp.Action == "expire" {
				break
			}
		}
	}
}

func (e *etcdLock) Release() error {
	path := path.Join(e.path, strings.Replace(e.id, "/", "-", -1))
	_, err := e.api.Delete(context.Background(), path, &client.DeleteOptions{
		PrevValue: e.node,
	})
	return err
}
