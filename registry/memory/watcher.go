package memory

import (
	"errors"

	"github.com/micro/go-micro/registry"
)

type memoryWatcher struct {
	id   string
	res  chan *registry.Result
	exit chan bool
}

func (m *memoryWatcher) Next() (*registry.Result, error) {
	select {
	case r := <-m.res:
		return r, nil
	case <-m.exit:
		return nil, errors.New("watcher stopped")
	}
}

func (m *memoryWatcher) Stop() {
	select {
	case <-m.exit:
		return
	default:
		close(m.exit)
	}
}
