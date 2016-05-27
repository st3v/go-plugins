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

// ListPods ...
func (c *client) ListPods(labels map[string]string) (*PodList, error) {
	var pods PodList
	err := api.NewRequest(c.opts).Get().Resource("pods").Params(&api.Params{LabelSelector: labels}).Do().Into(&pods)
	return &pods, err
}

// UpdatePod ...
func (c *client) UpdatePod(name string, p *Pod) (*Pod, error) {
	var pod Pod
	err := api.NewRequest(c.opts).Patch().Resource("pods").Name(name).Body(p).Do().Into(&pod)
	return &pod, err
}

// WatchPods ...
func (c *client) WatchPods(labels map[string]string) (watch.Watch, error) {
	return api.NewRequest(c.opts).Get().Resource("pods").Params(&api.Params{LabelSelector: labels}).Watch()
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
