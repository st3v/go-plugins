package zookeeper

import (
	"errors"
	"sync"
	"time"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
	"github.com/samuel/go-zookeeper/zk"

	hash "github.com/mitchellh/hashstructure"
)

var (
	prefix = "/micro-registry"
)

type zookeeperRegistry struct {
	client  *zk.Conn
	options registry.Options
	sync.Mutex
	register map[string]uint64
}

func init() {
	cmd.DefaultRegistries["zookeeper"] = NewRegistry
}

func (z *zookeeperRegistry) Deregister(s *registry.Service) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	// delete our hash of the service
	z.Lock()
	delete(z.register, s.Name)
	z.Unlock()

	for _, node := range s.Nodes {
		err := z.client.Delete(nodePath(s.Name, node.Id), -1)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *zookeeperRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	var options registry.RegisterOptions
	for _, o := range opts {
		o(&options)
	}

	// create hash of service; uint64
	h, err := hash.Hash(s, nil)
	if err != nil {
		return err
	}

	// get existing hash
	z.Lock()
	v, ok := z.register[s.Name]
	z.Unlock()

	// the service is unchanged, skip registering
	if ok && v == h {
		return nil
	}

	service := &registry.Service{
		Name:      s.Name,
		Version:   s.Version,
		Metadata:  s.Metadata,
		Endpoints: s.Endpoints,
	}

	for _, node := range s.Nodes {
		service.Nodes = []*registry.Node{node}
		e, _, err := z.client.Exists(nodePath(service.Name, node.Id))
		if err != nil {
			return err
		}

		if e {
			_, err := z.client.Set(nodePath(service.Name, node.Id), encode(service), -1)
			if err != nil {
				return err
			}
		} else {
			err := createPath(nodePath(service.Name, node.Id), encode(service), z.client)
			if err != nil {
				return err
			}
		}
	}

	// save our hash of the service
	z.Lock()
	z.register[s.Name] = h
	z.Unlock()

	return nil
}

func (z *zookeeperRegistry) GetService(name string) ([]*registry.Service, error) {
	l, _, err := z.client.Children(servicePath(name))
	if err != nil {
		return nil, err
	}

	serviceMap := map[string]*registry.Service{}

	for _, n := range l {
		_, stat, err := z.client.Children(nodePath(name, n))
		if err != nil {
			return nil, err
		}
		if stat.NumChildren > 0 {
			continue
		}

		b, _, err := z.client.Get(nodePath(name, n))
		if err != nil {
			return nil, err
		}
		sn := decode(b)

		s, ok := serviceMap[sn.Version]
		if !ok {
			s = &registry.Service{
				Name:      sn.Name,
				Version:   sn.Version,
				Metadata:  sn.Metadata,
				Endpoints: sn.Endpoints,
			}
			serviceMap[s.Version] = s
		}

		for _, node := range sn.Nodes {
			s.Nodes = append(s.Nodes, node)
		}
	}

	var services []*registry.Service
	for _, service := range serviceMap {
		services = append(services, service)
	}
	return services, nil
}

func (z *zookeeperRegistry) ListServices() ([]*registry.Service, error) {
	var services []*registry.Service
	serviceMap := map[string]*registry.Service{}

	err := getServices(z.client, serviceMap)
	if err != nil {
		return nil, err
	}
	for _, service := range serviceMap {
		services = append(services, service)
	}

	return services, nil
}

func (z *zookeeperRegistry) String() string {
	return "zookeeper"
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	var options registry.Options
	for _, o := range opts {
		o(&options)
	}

	if options.Timeout == 0 {
		options.Timeout = 5
	}

	var cAddrs []string
	for _, addr := range options.Addrs {
		if len(addr) == 0 {
			continue
		}
		cAddrs = append(cAddrs, addr)
	}

	if len(cAddrs) == 0 {
		cAddrs = []string{"127.0.0.1:2181"}
	}

	c, _, _ := zk.Connect(cAddrs, time.Second*options.Timeout)
	e := &zookeeperRegistry{
		client:   c,
		options:  options,
		register: make(map[string]uint64),
	}

	createPath(prefix, []byte{}, c)

	return e
}

func (z *zookeeperRegistry) Watch() (registry.Watcher, error) {
	return newZookeeperWatcher(z)
}
