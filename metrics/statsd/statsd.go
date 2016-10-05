package statsd

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/micro/go-os/metrics"
)

type statsd struct {
	exit chan bool
	opts metrics.Options

	sync.Mutex
	buf chan string
}

type counter struct {
	id  string
	buf chan string
}

type gauge struct {
	id  string
	buf chan string
}

type histogram struct {
	id  string
	buf chan string
}

var (
	maxBufferSize = 500
)

func newStatsd(opts ...metrics.Option) metrics.Metrics {
	options := metrics.Options{
		Namespace:     metrics.DefaultNamespace,
		BatchInterval: metrics.DefaultBatchInterval,
	}

	for _, o := range opts {
		o(&options)
	}

	s := &statsd{
		exit: make(chan bool),
		opts: options,
		buf:  make(chan string, 1000),
	}

	go s.run()
	return s
}

func (c *counter) format(v uint64) string {
	return fmt.Sprintf("%s:%d|c", c.id, v)
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
	return c
}

func (g *gauge) format(v int64) string {
	return fmt.Sprintf("%s:%d|g", g.id, v)
}

func (g *gauge) Set(d int64) {
	g.buf <- g.format(d)
}

func (g *gauge) Reset() {
	g.buf <- g.format(0)
}

func (g *gauge) WithFields(f metrics.Fields) metrics.Gauge {
	return g
}

func (h *histogram) format(v int64) string {
	return fmt.Sprintf("%s:%d|ms", h.id, v)
}

func (h *histogram) Record(d int64) {
	h.buf <- h.format(d)
}

func (h *histogram) Reset() {
	h.buf <- h.format(0)
}

func (h *histogram) WithFields(f metrics.Fields) metrics.Histogram {
	return h
}

func (s *statsd) run() {
	t := time.NewTicker(s.opts.BatchInterval)
	buf := bytes.NewBuffer(nil)

	conn, _ := net.DialTimeout("udp", s.opts.Collectors[0], time.Second)
	defer conn.Close()

	for {
		select {
		case <-s.exit:
			t.Stop()
			buf.Reset()
			buf = nil
			return
		case v := <-s.buf:
			buf.Write([]byte(fmt.Sprintf("%s.%s\n", s.opts.Namespace, v)))
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

func (s *statsd) Close() error {
	select {
	case <-s.exit:
		return nil
	default:
		close(s.exit)
	}
	return nil
}

func (s *statsd) Init(opts ...metrics.Option) error {
	for _, o := range opts {
		o(&s.opts)
	}
	return nil
}

func (s *statsd) Counter(id string) metrics.Counter {
	return &counter{
		id:  id,
		buf: s.buf,
	}
}

func (s *statsd) Gauge(id string) metrics.Gauge {
	return &gauge{
		id:  id,
		buf: s.buf,
	}
}

func (s *statsd) Histogram(id string) metrics.Histogram {
	return &histogram{
		id:  id,
		buf: s.buf,
	}
}

func (s *statsd) String() string {
	return "statsd"
}

func NewMetrics(opts ...metrics.Option) metrics.Metrics {
	return newStatsd(opts...)
}
