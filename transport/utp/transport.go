package utp

import (
	"bufio"
	"encoding/gob"

	"github.com/anacrolix/utp"
	"github.com/micro/go-micro/transport"
)

func (u *utpTransport) Dial(addr string, opts ...transport.DialOption) (transport.Client, error) {
	dopts := transport.DialOptions{
		Timeout: transport.DefaultDialTimeout,
	}

	for _, opt := range opts {
		opt(&dopts)
	}

	c, err := utp.DialTimeout(addr, dopts.Timeout)
	if err != nil {
		return nil, err
	}

	encBuf := bufio.NewWriter(c)

	return &utpClient{
		dialOpts: dopts,
		conn:     c,
		encBuf:   encBuf,
		enc:      gob.NewEncoder(encBuf),
		dec:      gob.NewDecoder(c),
		timeout:  u.opts.Timeout,
	}, nil
}

func (u *utpTransport) Listen(addr string, opts ...transport.ListenOption) (transport.Listener, error) {
	var options transport.ListenOptions
	for _, o := range opts {
		o(&options)
	}

	l, err := listen(addr, utp.Listen)
	if err != nil {
		return nil, err
	}

	return &utpListener{
		t:    u.opts.Timeout,
		l:    l,
		opts: options,
	}, nil
}

func (u *utpTransport) String() string {
	return "utp"
}
