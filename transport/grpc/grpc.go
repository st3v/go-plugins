// Package grpc provides a grpc transport
package grpc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/transport"
	mls "github.com/micro/misc/lib/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/micro/go-plugins/transport/grpc/proto"

	"golang.org/x/net/context"
)

type grpcTransport struct {
	opts transport.Options
}

type grpcTransportListener struct {
	listener net.Listener
	secure   bool
	tls      *tls.Config
}

func init() {
	cmd.DefaultTransports["grpc"] = NewTransport
}

func getTLSConfig(addr string) (*tls.Config, error) {
	hosts := []string{addr}

	// check if its a valid host:port
	if host, _, err := net.SplitHostPort(addr); err == nil {
		if len(host) == 0 {
			hosts = getIPAddrs()
		} else {
			hosts = []string{host}
		}
	}

	// generate a certificate
	cert, err := mls.Certificate(hosts...)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

func getIPAddrs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var ipAddrs []string

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue
			}

			ipAddrs = append(ipAddrs, ip.String())
		}
	}

	return ipAddrs
}

func listen(addr string, fn func(string) (net.Listener, error)) (net.Listener, error) {
	// host:port || host:min-max
	parts := strings.Split(addr, ":")

	//
	if len(parts) < 2 {
		return fn(addr)
	}

	// try to extract port range
	ports := strings.Split(parts[len(parts)-1], "-")

	// single port
	if len(ports) < 2 {
		return fn(addr)
	}

	// we have a port range

	// extract min port
	min, err := strconv.Atoi(ports[0])
	if err != nil {
		return nil, errors.New("unable to extract port range")
	}

	// extract max port
	max, err := strconv.Atoi(ports[1])
	if err != nil {
		return nil, errors.New("unable to extract port range")
	}

	// set host
	host := parts[:len(parts)-1]

	// range the ports
	for port := min; port <= max; port++ {
		// try bind to host:port
		ln, err := fn(fmt.Sprintf("%s:%d", host, port))
		if err == nil {
			return ln, nil
		}

		// hit max port
		if port == max {
			return nil, err
		}
	}

	// why are we here?
	return nil, fmt.Errorf("unable to bind to %s", addr)
}

func (t *grpcTransportListener) Addr() string {
	return t.listener.Addr().String()
}

func (t *grpcTransportListener) Close() error {
	return t.listener.Close()
}

func (t *grpcTransportListener) Accept(fn func(transport.Socket)) error {
	var opts []grpc.ServerOption

	// setup tls if specified
	if t.secure || t.tls != nil {
		config := t.tls
		if config == nil {
			var err error
			addr := t.listener.Addr().String()
			config, err = getTLSConfig(addr)
			if err != nil {
				return err
			}
		}

		creds := credentials.NewTLS(config)
		opts = append(opts, grpc.Creds(creds))
	}

	// new service
	srv := grpc.NewServer(opts...)

	// register service
	pb.RegisterTransportServer(srv, &microTransport{fn: fn})

	// start serving
	return srv.Serve(t.listener)
}

func (t *grpcTransport) Dial(addr string, opts ...transport.DialOption) (transport.Client, error) {
	dopts := transport.DialOptions{
		Timeout: transport.DefaultDialTimeout,
	}

	for _, opt := range opts {
		opt(&dopts)
	}

	options := []grpc.DialOption{
		grpc.WithTimeout(dopts.Timeout),
	}

	if t.opts.Secure || t.opts.TLSConfig != nil {
		config := t.opts.TLSConfig
		if config == nil {
			config = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		creds := credentials.NewTLS(config)
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	// dial the server
	conn, err := grpc.Dial(addr, options...)
	if err != nil {
		return nil, err
	}

	// create stream
	stream, err := pb.NewTransportClient(conn).Stream(context.Background())
	if err != nil {
		return nil, err
	}

	// return a client
	return &grpcTransportClient{
		conn:   conn,
		stream: stream,
	}, nil
}

func (t *grpcTransport) Listen(addr string, opts ...transport.ListenOption) (transport.Listener, error) {
	var options transport.ListenOptions
	for _, o := range opts {
		o(&options)
	}

	ln, err := listen(addr, func(addr string) (net.Listener, error) {
		return net.Listen("tcp", addr)
	})
	if err != nil {
		return nil, err
	}

	return &grpcTransportListener{
		listener: ln,
		tls:      t.opts.TLSConfig,
		secure:   t.opts.Secure,
	}, nil
}

func (t *grpcTransport) String() string {
	return "grpc"
}

func NewTransport(opts ...transport.Option) transport.Transport {
	var options transport.Options
	for _, o := range opts {
		o(&options)
	}
	return &grpcTransport{opts: options}
}
