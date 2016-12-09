package telegraf

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-os/metrics"
)

type telegraf struct {
	exit chan bool
	opts metrics.Options

	sync.Mutex
	buf chan string
}

type counter struct {
	id  string
	buf chan string
	f   metrics.Fields
}

type gauge struct {
	id  string
	buf chan string
	f   metrics.Fields
}

type histogram struct {
	id  string
	buf chan string
	f   metrics.Fields
}

var (
	maxBufferSize = 500
)

func (c *counter) format(v uint64) string {
	keys := []string{c.id}

	for k, v := range c.f {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s:%d|c", strings.Join(keys, ","), v)
}

func (c *counter) Incr(d uint64) {
	c.buf <- c.format(d)
}

func (c *counter) Decr(d uint64) {
	c.buf <- c.format(-d)
}

func (c *counter) Reset() {
	c.buf <- c.format(0)
}

func (c *counter) WithFields(f metrics.Fields) metrics.Counter {
	nf := make(metrics.Fields)

	for k, v := range c.f {
		nf[k] = v
	}

	for k, v := range f {
		nf[k] = v
	}

	return &counter{
		buf: c.buf,
		id:  c.id,
		f:   nf,
	}
}

func (g *gauge) format(v int64) string {
	keys := []string{g.id}

	for k, v := range g.f {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s:%d|g", strings.Join(keys, ","), v)
}

func (g *gauge) Set(d int64) {
	g.buf <- g.format(d)
}

func (g *gauge) Reset() {
	g.buf <- g.format(0)
}

func (g *gauge) WithFields(f metrics.Fields) metrics.Gauge {
	nf := make(metrics.Fields)

	for k, v := range g.f {
		nf[k] = v
	}

	for k, v := range f {
		nf[k] = v
	}

	return &gauge{
		buf: g.buf,
		id:  g.id,
		f:   nf,
	}
}

func (h *histogram) format(v int64) string {
	keys := []string{h.id}

	for k, v := range h.f {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s:%d|ms", strings.Join(keys, ","), v)
}

func (h *histogram) Record(d int64) {
	h.buf <- h.format(d)
}

func (h *histogram) Reset() {
	h.buf <- h.format(0)
}

func (h *histogram) WithFields(f metrics.Fields) metrics.Histogram {
	nf := make(metrics.Fields)

	for k, v := range h.f {
		nf[k] = v
	}

	for k, v := range f {
		nf[k] = v
	}

	return &histogram{
		buf: h.buf,
		id:  h.id,
		f:   nf,
	}
}

func (t *telegraf) run() {
	tc := time.NewTicker(t.opts.BatchInterval)
	buf := bytes.NewBuffer(nil)

	conn, _ := net.DialTimeout("udp", t.opts.Collectors[0], time.Second)
	defer conn.Close()

	for {
		select {
		case <-t.exit:
			tc.Stop()
			buf.Reset()
			buf = nil
			return
		case v := <-t.buf:
			buf.Write([]byte(fmt.Sprintf("%s.%s\n", t.opts.Namespace, v)))
			if buf.Len() < maxBufferSize {
				continue
			}
			conn.Write(buf.Bytes())
			buf.Reset()
		case <-tc.C:
			conn.Write(buf.Bytes())
			buf.Reset()
		}
	}
}

func (t *telegraf) Close() error {
	select {
	case <-t.exit:
		return nil
	default:
		close(t.exit)
	}
	return nil
}

func (t *telegraf) Init(opts ...metrics.Option) error {
	for _, o := range opts {
		o(&t.opts)
	}
	return nil
}

func (t *telegraf) Counter(id string) metrics.Counter {
	return &counter{
		id:  id,
		buf: t.buf,
		f:   t.opts.Fields,
	}
}

func (t *telegraf) Gauge(id string) metrics.Gauge {
	return &gauge{
		id:  id,
		buf: t.buf,
		f:   t.opts.Fields,
	}
}

func (t *telegraf) Histogram(id string) metrics.Histogram {
	return &histogram{
		id:  id,
		buf: t.buf,
		f:   t.opts.Fields,
	}
}

func (t *telegraf) String() string {
	return "telegraf"
}

func NewMetrics(opts ...metrics.Option) metrics.Metrics {
	options := metrics.Options{
		Namespace:     metrics.DefaultNamespace,
		BatchInterval: metrics.DefaultBatchInterval,
		Collectors:    []string{"127.0.0.1:8125"},
		Fields:        make(metrics.Fields),
	}

	for _, o := range opts {
		o(&options)
	}

	t := &telegraf{
		exit: make(chan bool),
		opts: options,
		buf:  make(chan string, 1000),
	}

	go t.run()
	return t
}
