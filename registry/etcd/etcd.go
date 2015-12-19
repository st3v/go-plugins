package etcd

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	etcd "github.com/coreos/etcd/client"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
	"golang.org/x/net/context"
)

var (
	prefix = "/micro-registry"
)

type etcdRegistry struct {
	client etcd.KeysAPI
	options registry.Options
}

func init() {
	cmd.Registries["etcd"] = NewRegistry
}

func encode(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds string) *registry.Service {
	var s *registry.Service
	json.Unmarshal([]byte(ds), &s)
	return s
}

func nodePath(s, id string) string {
	service := strings.Replace(s, "/", "-", -1)
	node := strings.Replace(id, "/", "-", -1)
	return filepath.Join(prefix, service, node)
}

func servicePath(s string) string {
	return filepath.Join(prefix, strings.Replace(s, "/", "-", -1))
}

func (e *etcdRegistry) Deregister(s *registry.Service) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	for _, node := range s.Nodes {
		_, err := e.client.Delete(ctx, nodePath(s.Name, node.Id), &etcd.DeleteOptions{Recursive: false})
		if err != nil {
			return err
		}
	}

	e.client.Delete(ctx, servicePath(s.Name), &etcd.DeleteOptions{Dir: true})
	return nil
}

func (e *etcdRegistry) Register(s *registry.Service) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	service := &registry.Service{
		Name:      s.Name,
		Version:   s.Version,
		Metadata:  s.Metadata,
		Endpoints: s.Endpoints,
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	_, err := e.client.Set(ctx, servicePath(s.Name), "", &etcd.SetOptions{PrevExist: etcd.PrevIgnore, Dir: true})
	if err != nil && !strings.HasPrefix(err.Error(), "102: Not a file") {
		return err
	}

	for _, node := range s.Nodes {
		service.Nodes = []*registry.Node{node}
		_, err := e.client.Set(ctx, nodePath(service.Name, node.Id), encode(service), &etcd.SetOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *etcdRegistry) GetService(name string) ([]*registry.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	rsp, err := e.client.Get(ctx, servicePath(name), &etcd.GetOptions{})
	if err != nil && !strings.HasPrefix(err.Error(), "100: Key not found") {
		return nil, err
	}

	serviceMap := map[string]*registry.Service{}

	for _, n := range rsp.Node.Nodes {
		if n.Dir {
			continue
		}
		sn := decode(n.Value)

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

func (e *etcdRegistry) ListServices() ([]*registry.Service, error) {
	var services []*registry.Service

	ctx, cancel := context.WithTimeout(context.Background(), e.options.Timeout)
	defer cancel()

	rsp, err := e.client.Get(ctx, prefix, &etcd.GetOptions{Recursive: true, Sort: true})
	if err != nil && !strings.HasPrefix(err.Error(), "100: Key not found") {
		return nil, err
	}

	if rsp == nil {
		return []*registry.Service{}, nil
	}

	for _, node := range rsp.Node.Nodes {
		service := &registry.Service{}
		for _, n := range node.Nodes {
			i := decode(n.Value)
			service.Name = i.Name
		}
		services = append(services, service)
	}

	return services, nil
}

func (e *etcdRegistry) Watch() (registry.Watcher, error) {
	return newEtcdWatcher(e)
}

func NewRegistry(addrs []string, opts ...registry.Option) registry.Registry {
        var opt registry.Options
        for _, o := range opts {
                o(&opt)
        }

	if opt.Timeout == 0 {
		opt.Timeout = etcd.DefaultRequestTimeout
	}

	var cAddrs []string

	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		cAddrs = append(cAddrs, addr)
	}

	if len(cAddrs) == 0 {
		cAddrs = []string{"http://127.0.0.1:2379"}
	}

	c, _ := etcd.New(etcd.Config{
		Endpoints: cAddrs,
	})

	e := &etcdRegistry{
		client: etcd.NewKeysAPI(c),
		options: opt,
	}

	return e
}
