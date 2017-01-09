package utp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

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
