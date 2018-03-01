package datadog

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-os/metrics"
)

type datadog struct {
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

func newMetrics(opts ...metrics.Option) metrics.Metrics {
	options := metrics.Options{
		Namespace:     metrics.DefaultNamespace,
		BatchInterval: metrics.DefaultBatchInterval,
		Fields:        make(metrics.Fields),
	}

	for _, o := range opts {
		o(&options)
	}

	s := &datadog{
		exit: make(chan bool),
		opts: options,
		buf:  make(chan string, 1000),
	}

	go s.run()
	return s
}

func (c *counter) format(v uint64) string {
	return fmt.Sprintf("%s:%d|c%s", c.id, v, formatTags(c.f))
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
		id:  c.id,
		buf: c.buf,
		f:   nf,
	}
}

func (g *gauge) format(v int64) string {
	return fmt.Sprintf("%s:%d|g%s", g.id, v, formatTags(g.f))
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
		id:  g.id,
		buf: g.buf,
		f:   nf,
	}
}

func (h *histogram) format(v int64) string {
	return fmt.Sprintf("%s:%d|ms%s", h.id, v, formatTags(h.f))
}

func formatTags(f metrics.Fields) string {
	s := ""

	if len(f) > 0 {
		tags := []string{}

		for k, v := range f {
			if v != "" {
				k = fmt.Sprintf("%s:%s", k, v)
			}
			tags = append(tags, k)
		}

		if len(tags) > 0 {
			s = fmt.Sprintf("%s|#%s", s, strings.Join(tags, ","))
		}
	}

	return s
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
		id:  h.id,
		buf: h.buf,
		f:   nf,
	}
}

func (d *datadog) run() {
	t := time.NewTicker(d.opts.BatchInterval)
	buf := bytes.NewBuffer(nil)

	conn, _ := net.DialTimeout("udp", d.opts.Collectors[0], time.Second)
	defer conn.Close()

	for {
		select {
		case <-d.exit:
			t.Stop()
			buf.Reset()
			buf = nil
			return
		case v := <-d.buf:
			buf.Write([]byte(fmt.Sprintf("%s.%s\n", d.opts.Namespace, v)))
			if buf.Len() < maxBufferSize {
				continue
			}
			conn.Write(buf.Bytes())
			buf.Reset()
		case <-t.C:
			conn.Write(buf.Bytes())
			buf.Reset()
		}
	}
}

func (d *datadog) Close() error {
	select {
	case <-d.exit:
		return nil
	default:
		close(d.exit)
	}
	return nil
}

func (d *datadog) Init(opts ...metrics.Option) error {
	for _, o := range opts {
		o(&d.opts)
	}
	return nil
}

func (d *datadog) Counter(id string) metrics.Counter {
	return &counter{
		id:  id,
		buf: d.buf,
		f:   d.opts.Fields,
	}
}

func (d *datadog) Gauge(id string) metrics.Gauge {
	return &gauge{
		id:  id,
		buf: d.buf,
		f:   d.opts.Fields,
	}
}

func (d *datadog) Histogram(id string) metrics.Histogram {
	return &histogram{
		id:  id,
		buf: d.buf,
		f:   d.opts.Fields,
	}
}

func (d *datadog) String() string {
	return "datadog"
}

func NewMetrics(opts ...metrics.Option) metrics.Metrics {
	return newMetrics(opts...)
}
