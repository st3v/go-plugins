package gossip

import (
	"errors"

	"github.com/micro/go-micro/registry"
)

type gossipWatcher struct {
	next chan *registry.Result
	stop chan bool
}

func newGossipWatcher(ch chan *registry.Result, exit chan bool) (registry.Watcher, error) {
	stop := make(chan bool)

	go func() {
		<-stop
		close(exit)
	}()

	return &gossipWatcher{
		next: ch,
		stop: stop,
	}, nil
}

func (m *gossipWatcher) Next() (*registry.Result, error) {
	select {
	case r, ok := <-m.next:
		if !ok {
			return nil, errors.New("result chan closed")
		}
		return r, nil
	case <-m.stop:
		return nil, errors.New("watcher stopped")
	}
}

func (m *gossipWatcher) Stop() {
	select {
	case <-m.stop:
		return
	default:
		close(m.stop)
	}
}
