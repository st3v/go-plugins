package nats

import (
	"encoding/json"
	"time"

	"github.com/micro/go-micro/registry"
	"github.com/nats-io/nats"
)

type natsWatcher struct {
	sub *nats.Subscription
}

func (n *natsWatcher) Next() (*registry.Result, error) {
	var result *registry.Result
	for {
		m, err := n.sub.NextMsg(time.Minute)
		if err != nil && err == nats.ErrTimeout {
			continue
		} else if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(m.Data, &result); err != nil {
			return nil, err
		}
		break
	}
	return result, nil
}

func (n *natsWatcher) Stop() {
	n.sub.Unsubscribe()
}
