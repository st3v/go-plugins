package etcd

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/micro/go-log"
	"github.com/divisionone/go-micro/registry"
	"github.com/micro/go-os/sync"
	"github.com/pborman/uuid"
)

type etcdSync struct {
	opts   sync.Options
	client client.Client
}

func (e *etcdSync) Lock(id string, opts ...sync.LockOption) (sync.Lock, error) {
	var options sync.LockOptions
	for _, o := range opts {
		o(&options)
	}

	return &etcdLock{
		path: e.opts.Namespace,
		id:   id,
		node: fmt.Sprintf("%s-%d", uuid.NewUUID().String(), time.Now().UnixNano()),
		api:  client.NewKeysAPI(e.client),
		opts: options,
	}, nil
}

func (e *etcdSync) Leader(id string, opts ...sync.LeaderOption) (sync.Leader, error) {
	return &etcdLeader{
		path: e.opts.Namespace,
		id:   id,
		node: &registry.Node{Id: uuid.NewUUID().String()},
		api:  client.NewKeysAPI(e.client),
	}, nil
}

func (e *etcdSync) String() string {
	return "etcd"
}

func NewSync(opts ...sync.Option) sync.Sync {
	options := sync.Options{
		Namespace: sync.DefaultNamespace,
	}

	for _, o := range opts {
		o(&options)
	}

	var endpoints []string

	for _, addr := range options.Nodes {
		if len(addr) > 0 {
			endpoints = append(endpoints, addr)
		}
	}

	if len(endpoints) == 0 {
		endpoints = []string{"http://127.0.0.1:2379"}
	}

	// TODO: parse addresses
	c, err := client.New(client.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &etcdSync{
		client: c,
		opts:   options,
	}
}
