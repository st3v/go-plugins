package consul

import (
	"encoding/json"
	"errors"
	mtx "sync"

	"github.com/hashicorp/consul/api"
	"github.com/micro/go-micro/registry"
	sync "github.com/micro/go-os/sync"
)

type consulLeader struct {
	opts sync.LeaderOptions
	c    *api.Client

	id  string
	key string
	// marshalled registry node
	srv []byte

	// used to sync back status
	statusCh chan sync.LeaderStatus

	mtx.Mutex
	status sync.LeaderStatus
}

type consulElected struct {
	ch chan sync.LeaderStatus
	rv <-chan struct{}
	l  *api.Lock
}

func (c *consulLeader) Leader() (*registry.Node, error) {
	kv, _, err := c.c.KV().Get(c.key, nil)
	if err != nil {
		return nil, err
	}
	var node *registry.Node
	if err := json.Unmarshal(kv.Value, &node); err != nil {
		return nil, err
	}
	return node, nil
}

func (c *consulLeader) Id() string {
	return c.id
}

func (c *consulLeader) Elect() (sync.Elected, error) {
	lc, err := c.c.LockOpts(&api.LockOptions{
		Key:   c.key,
		Value: c.srv,
	})
	if err != nil {
		return nil, err
	}

	rv, err := lc.Lock(nil)
	if err != nil {
		return nil, err
	}

	c.statusCh <- sync.ElectedStatus

	// lock acquired
	return &consulElected{
		rv: rv,
		l:  lc,
		ch: c.statusCh,
	}, nil
}

func (c *consulLeader) Status() (sync.LeaderStatus, error) {
	c.Lock()
	defer c.Unlock()
	return c.status, nil
}

func (c *consulElected) Revoked() (chan struct{}, error) {
	select {
	case <-c.rv:
		return nil, errors.New("already revoked")
	default:
	}

	ch := make(chan struct{}, 1)

	go func() {
		st := <-ch
		c.ch <- sync.FollowerStatus
		ch <- st
	}()

	return ch, nil
}

func (c *consulElected) Resign() error {
	c.ch <- sync.FollowerStatus
	err := c.l.Unlock()
	return err
}
