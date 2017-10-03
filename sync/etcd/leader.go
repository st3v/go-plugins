package etcd

import (
	"encoding/json"
	"errors"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/micro/go-micro/registry"
	psync "github.com/micro/go-os/sync"

	"golang.org/x/net/context"
)

type etcdLeader struct {
	id   string
	path string
	node *registry.Node
	api  client.KeysAPI

	sync.Mutex
	status psync.LeaderStatus
}

type etcdElected struct {
	resign chan bool
	revoke chan struct{}
}

func (e *etcdLeader) Id() string {
	return e.id
}

func (e *etcdLeader) elect(val string, opts *client.SetOptions, wait bool) error {
	// make path
	path := path.Join(e.path, strings.Replace(e.id, "/", "-", -1))
	ctx := context.Background()

	for {
		// try set the lock
		_, err := e.api.Set(ctx, path, val, opts)
		if err == nil {
			return nil
		}

		// should we wait?
		if !wait {
			return err
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

func (e *etcdLeader) resign(val string) {
	path := path.Join(e.path, strings.Replace(e.id, "/", "-", -1))
	e.api.Delete(context.Background(), path, &client.DeleteOptions{
		PrevValue: val,
	})
}

func (e *etcdLeader) loop(val string, ee *etcdElected) {
	t := time.NewTicker(time.Second * 30)

	defer func() {
		t.Stop()
		e.Lock()
		e.status = psync.FollowerStatus
		e.Unlock()
	}()

	for {
		select {
		// re-elect self on the tick
		case <-t.C:
			// attempt to re-elect
			err := e.elect(val, &client.SetOptions{
				PrevValue: val,
				TTL:       time.Minute,
			}, false)

			// return on error
			if err != nil {
				select {
				case <-ee.revoke:
					return
				default:
					close(ee.revoke)
				}
				return
			}
		case <-ee.revoke:
			return
		case <-ee.resign:
			e.resign(val)
			return
		}
	}
}

func (e *etcdLeader) Leader() (*registry.Node, error) {
	path := path.Join(e.path, strings.Replace(e.id, "/", "-", -1))

	rsp, err := e.api.Get(context.Background(), path, nil)
	if err == nil {
		var node *registry.Node
		if err := json.Unmarshal([]byte(rsp.Node.Value), &node); err != nil {
			return nil, err
		}
		return node, nil
	}

	errr, ok := err.(client.Error)
	if ok && errr.Code == client.ErrorCodeKeyNotFound {
		return nil, errors.New("no leader")
	}

	return nil, err
}

func (e *etcdLeader) Elect() (psync.Elected, error) {
	// set candidate status
	e.Lock()
	if e.status == psync.ElectedStatus {
		e.Unlock()
		return nil, errors.New("already the elected")
	}
	e.status = psync.CandidateStatus
	e.Unlock()

	// marshal self
	b, err := json.Marshal(e.node)
	if err != nil {
		return nil, err
	}
	val := string(b)

	// attempt to elect
	if err := e.elect(val, &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       time.Minute,
	}, true); err != nil {
		return nil, err
	}

	// set elected status
	e.Lock()
	e.status = psync.ElectedStatus
	e.Unlock()

	ee := &etcdElected{
		resign: make(chan bool),
		revoke: make(chan struct{}),
	}

	// loop for maintaining leader
	go e.loop(val, ee)

	return ee, nil
}

func (e *etcdLeader) Status() (psync.LeaderStatus, error) {
	e.Lock()
	defer e.Unlock()
	return e.status, nil
}

func (e *etcdElected) Revoked() (chan struct{}, error) {
	return e.revoke, nil
}

func (e *etcdElected) Resign() error {
	select {
	case <-e.resign:
		return nil
	default:
		close(e.resign)
	}
	return nil
}
