package client

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/micro/go-plugins/registry/kubernetes/client/api"
	"github.com/micro/go-plugins/registry/kubernetes/client/watch"
)

// Client ...
type client struct {
	opts *api.Options
}

// UpdatePod ...
func (c *client) UpdatePod(name string, p *Pod) (*Pod, error) {
	var pod Pod
	err := api.NewRequest(c.opts).Patch().Resource("pods").Name(name).Body(p).Do().Into(&pod)
	return &pod, err
}

// CreateService ...
func (c *client) CreateService(s *Service) (*Service, error) {
	var service Service
	err := api.NewRequest(c.opts).Post().Resource("services").Body(s).Do().Into(&service)
	return &service, err
}

// UpdateService ...
func (c *client) UpdateService(name string, s *Service) (*Service, error) {
	var service Service
	err := api.NewRequest(c.opts).Patch().Resource("services").Name(name).Body(s).Do().Into(&service)
	return &service, err
}

// ListEndpoints ...
func (c *client) ListEndpoints(labels map[string]string) (*EndpointsList, error) {
	var endpoints EndpointsList
	err := api.NewRequest(c.opts).Get().Resource("endpoints").Params(&api.Params{LabelSelector: labels}).Do().Into(&endpoints)
	return &endpoints, err
}

// GetEndpoints ...
func (c *client) GetEndpoints(name string) (*Endpoints, error) {
	var endpoints Endpoints
	err := api.NewRequest(c.opts).Get().Resource("endpoints").Name(name).Do().Into(&endpoints)
	return &endpoints, err
}

// UpdateEndpoints ...
func (c *client) UpdateEndpoints(name string, s *Endpoints) (*Endpoints, error) {
	var endpoints Endpoints
	err := api.NewRequest(c.opts).Patch().Resource("endpoints").Name(name).Body(s).Do().Into(&endpoints)
	return &endpoints, err
}

// WatchEndpoints ...
func (c *client) WatchEndpoints(labels map[string]string) (watch.Watch, error) {
	return api.NewRequest(c.opts).Get().Resource("endpoints").Params(&api.Params{LabelSelector: labels}).Watch()
}

// NewClientByHost sets up a client by host
func NewClientByHost(host string) Kubernetes {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{},
		DisableCompression: true,
	}

	c := &http.Client{
		Transport: tr,
	}

	return &client{
		opts: &api.Options{
			Client:    c,
			Host:      host,
			Namespace: "default",
		},
	}
}

// NewClientInCluster should work similarily to the official api
// NewInClient by setting up a client configuration for use within
// a k8s pod.
func NewClientInCluster() Kubernetes {
	host := "https://" + os.Getenv("KUBERNETES_SERVICE_HOST") + ":" + os.Getenv("KUBERNETES_SERVICE_PORT")
	sa := "/var/run/secrets/kubernetes.io/serviceaccount"

	s, err := os.Stat(sa)
	if err != nil {
		log.Fatal(err)
	}
	if s == nil || !s.IsDir() {
		log.Fatal(errors.New("no k8s service account found"))
	}

	token, err := ioutil.ReadFile(path.Join(sa, "token"))
	if err != nil {
		log.Fatal(err)
	}
	t := string(token)

	crt, err := CertPoolFromFile(path.Join(sa, "ca.crt"))
	if err != nil {
		log.Fatal(err)
	}

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: crt,
			},
			DisableCompression: true,
		},
	}

	return &client{
		opts: &api.Options{
			Client:      c,
			Host:        host,
			Namespace:   "default",
			BearerToken: &t,
		},
	}
}
