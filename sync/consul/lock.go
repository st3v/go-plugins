package consul

import (
	"errors"

	"github.com/hashicorp/consul/api"
	sync "github.com/micro/go-os/sync"
)

type consulLock struct {
	opts sync.LockOptions
	l    *api.Lock
	id   string
}

func (c *consulLock) Id() string {
	return c.id
}

func (c *consulLock) Acquire() error {
	lc, err := c.l.Lock(nil)
	if err != nil {
		return err
	}

	select {
	case <-lc:
		return errors.New("lock lost")
	default:
	}

	return nil
}

func (c *consulLock) Release() error {
	return c.l.Unlock()
}
