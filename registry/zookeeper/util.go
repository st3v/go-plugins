package zookeeper

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/micro/go-micro/registry"
	"github.com/samuel/go-zookeeper/zk"
)

func encode(s *registry.Service) []byte {
	b, _ := json.Marshal(s)
	return b
}

func decode(ds []byte) *registry.Service {
	var s *registry.Service
	json.Unmarshal(ds, &s)
	return s
}

func nodePath(s, id string) string {
	service := strings.Replace(s, "/", "-", -1)
	node := strings.Replace(id, "/", "-", -1)
	return path.Join(prefix, service, node)
}

func servicePath(s string) string {
	return path.Join(prefix, strings.Replace(s, "/", "-", -1))
}

func createPath(path string, data []byte, client *zk.Conn) error {
	var err error
	name := "/"
	p := strings.Split(path, "/")

	for _, v := range p[1 : len(p)-1] {
		name += v
		e, _, _ := client.Exists(name)
		if !e {
			_, err = client.Create(name, []byte{}, int32(0), zk.WorldACL(zk.PermAll))
			if err != nil {
				return err
			}
		}
		name += "/"
	}
	_, err = client.Create(path, data, int32(0), zk.WorldACL(zk.PermAll))
	return err
}

func getServices(c *zk.Conn, vars map[string]*registry.Service) error {
	services, _, err := c.Children(prefix)
	if err != nil {
		return err
	}

	for _, key := range services {
		s := servicePath(key)
		nodes, _, err := c.Children(s)
		if err != nil {
			return err
		}

		for _, node := range nodes {
			_, stat, err := c.Children(nodePath(key, node))
			if err != nil {
				return err
			}
			if stat.NumChildren == 0 {
				b, _, err := c.Get(nodePath(key, node))
				if err != nil {
					return err
				}
				service := &registry.Service{}
				i := decode(b)
				service.Name = i.Name
				vars[s] = service
			}
		}
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
