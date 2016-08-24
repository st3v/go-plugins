package eureka

/*
	Eureka is a plugin for Netflix Eureka service discovery
*/

import (
	"crypto/tls"
	"time"

	"golang.org/x/net/context"

	"github.com/st3v/go-eureka"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
)

type eurekaClient interface {
	Register(*eureka.Instance) error
	Deregister(*eureka.Instance) error
	Heartbeat(*eureka.Instance) error
	Apps() ([]*eureka.App, error)
	App(appName string) (*eureka.App, error)
	AppInstance(appName, instanceID string) (*eureka.Instance, error)
	Watch(pollInterval time.Duration) *eureka.Watcher
}

type eurekaRegistry struct {
	client       eurekaClient
	opts         registry.Options
	pollInterval time.Duration
}

func init() {
	cmd.DefaultRegistries["eureka"] = NewRegistry
}

func newRegistry(opts ...registry.Option) registry.Registry {
	options := registry.Options{
		Context: context.Background(),
		Secure:  true,
	}

	for _, o := range opts {
		o(&options)
	}

	var cAddrs []string
	for _, addr := range options.Addrs {
		if len(addr) == 0 {
			continue
		}
		cAddrs = append(cAddrs, addr)
	}

	if len(cAddrs) == 0 {
		cAddrs = []string{"http://localhost:8080/eureka/v2"}
	}

	clientOpts := []eureka.Option{}
	if creds, ok := options.Context.Value(contextOauth2Credentials{}).(oauth2Credentials); ok {
		clientOpts = append(clientOpts, eureka.Oauth2ClientCredentials(
			creds.ClientID,
			creds.ClientSecret,
			creds.TokenURL,
		))
	}

	if !options.Secure {
		if options.TLSConfig == nil {
			options.TLSConfig = new(tls.Config)
		}
		options.TLSConfig.InsecureSkipVerify = true
	}

	if options.TLSConfig != nil {
		clientOpts = append(clientOpts, eureka.TLSConfig(options.TLSConfig))
	}

	return &eurekaRegistry{
		client:       eureka.NewClient(cAddrs, clientOpts...),
		opts:         options,
		pollInterval: time.Second * 5,
	}
}

func (e *eurekaRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	instance, err := serviceToInstance(s)
	if err != nil {
		return err
	}

	if e.instanceRegistered(instance) {
		return e.client.Heartbeat(instance)
	}

	return e.client.Register(instance)
}

func (e *eurekaRegistry) Deregister(s *registry.Service) error {
	instance, err := serviceToInstance(s)
	if err != nil {
		return err
	}
	return e.client.Deregister(instance)
}

func (e *eurekaRegistry) GetService(name string) ([]*registry.Service, error) {
	app, err := e.client.App(name)
	if err != nil {
		return nil, err
	}
	return appToService(app), nil
}

func (e *eurekaRegistry) ListServices() ([]*registry.Service, error) {
	var services []*registry.Service

	apps, err := e.client.Apps()
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		services = append(services, appToService(app)...)
	}

	return services, nil
}

func (e *eurekaRegistry) Watch() (registry.Watcher, error) {
	return newWatcher(e.client), nil
}

func (e *eurekaRegistry) String() string {
	return "eureka"
}

func (e *eurekaRegistry) instanceRegistered(instance *eureka.Instance) bool {
	_, err := e.client.AppInstance(instance.AppName, instance.ID)
	return err == nil
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	return newRegistry(opts...)
}
