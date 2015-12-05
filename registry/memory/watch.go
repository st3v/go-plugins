package memory

import (
	"errors"
	"sync"

	"github.com/micro/go-micro/registry"
)

type memoryWatcher struct {
	once sync.Once
	next chan *registry.Result
	stop chan bool
}

func newMemoryWatcher(ch chan *registry.Result, exit chan bool) (registry.Watcher, error) {
	stop := make(chan bool)
	var once sync.Once

	go func() {
		<-stop
		close(exit)
	}()

	return &memoryWatcher{
		once: once,
		next: ch,
		stop: stop,
	}, nil
}

func (m *memoryWatcher) Next() (*registry.Result, error) {
	r, ok := <-m.next
	if !ok {
		return nil, errors.New("result chan closed")
	}
	return r, nil
}

func (m *memoryWatcher) Stop() {
	m.once.Do(func() {
		close(m.stop)
	})
}
