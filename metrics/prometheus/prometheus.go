package prometheus

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/micro/go-os/metrics"
	pr "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type prometheus struct {
	exit chan bool
	opts metrics.Options

	col []pr.Collector
	buf chan pr.Collector
}

type counter struct {
	id string
	cv *pr.CounterVec
	f  metrics.Fields
}

type gauge struct {
	id string
	gv *pr.GaugeVec
	f  metrics.Fields
}

type histogram struct {
	id string
	hv *pr.HistogramVec
	f  metrics.Fields
}

func newMetrics(opts ...metrics.Option) metrics.Metrics {
	options := metrics.Options{
		Namespace:     metrics.DefaultNamespace,
		BatchInterval: metrics.DefaultBatchInterval,
		Collectors:    []string{"http://127.0.0.1:9091"},
		Fields:        make(metrics.Fields),
	}

	for _, o := range opts {
		o(&options)
	}

	p := &prometheus{
		exit: make(chan bool),
		opts: options,
		buf:  make(chan pr.Collector, 1000),
	}

	go p.run()
	return p
}

func format(s string) string {
	return strings.Replace(s, ".", "_", -1)
}

func (c *counter) Incr(d uint64) {
	c.cv.With(pr.Labels(c.f)).Add(float64(d))
}

func (c *counter) Decr(d uint64) {
	c.cv.With(pr.Labels(c.f)).Add(-float64(d))
}

func (c *counter) Reset() {
	c.cv.With(pr.Labels(c.f)).Set(0.0)
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
		cv: c.cv,
		id: c.id,
		f:  nf,
	}
}

func (g *gauge) Set(d int64) {
	g.gv.With(pr.Labels(g.f)).Set(float64(d))
}

func (g *gauge) Reset() {
	g.gv.With(pr.Labels(g.f)).Set(0.0)
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
		gv: g.gv,
		id: g.id,
		f:  nf,
	}
}

func (h *histogram) Record(d int64) {
	h.hv.With(pr.Labels(h.f)).Observe(float64(d))
}

func (h *histogram) Reset() {
	h.hv.With(pr.Labels(h.f)).Observe(0.0)
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
		hv: h.hv,
		id: h.id,
		f:  nf,
	}
}

func (p *prometheus) run() {
	t := time.NewTicker(p.opts.BatchInterval)
	host, _ := os.Hostname()

	for {
		select {
		case <-p.exit:
			t.Stop()
			return
		case c := <-p.buf:
			p.col = append(p.col, c)
		case <-t.C:
			if err := push.AddCollectors(p.opts.Namespace, map[string]string{"host": host}, p.opts.Collectors[0], p.col...); err != nil {
				log.Print(p.String()+": ", err)
			}
		}
	}
}

func (p *prometheus) Close() error {
	select {
	case <-p.exit:
		return nil
	default:
		close(p.exit)
	}
	return nil
}

func (p *prometheus) Init(opts ...metrics.Option) error {
	for _, o := range opts {
		o(&p.opts)
	}
	return nil
}

func (p *prometheus) Counter(id string) metrics.Counter {
	var fields []string
	for k, _ := range p.opts.Fields {
		fields = append(fields, k)
	}

	cv := pr.NewCounterVec(pr.CounterOpts{
		Namespace: format(p.opts.Namespace),
		Name:      format(id),
		Help:      "counter",
	}, fields)

	p.buf <- cv

	return &counter{
		id: id,
		cv: cv,
		f:  p.opts.Fields,
	}
}

func (p *prometheus) Gauge(id string) metrics.Gauge {
	var fields []string
	for k, _ := range p.opts.Fields {
		fields = append(fields, k)
	}

	gv := pr.NewGaugeVec(pr.GaugeOpts{
		Namespace: format(p.opts.Namespace),
		Name:      format(id),
		Help:      "gauge",
	}, fields)

	p.buf <- gv

	return &gauge{
		id: id,
		gv: gv,
		f:  p.opts.Fields,
	}
}

func (p *prometheus) Histogram(id string) metrics.Histogram {
	var fields []string
	for k, _ := range p.opts.Fields {
		fields = append(fields, k)
	}

	hv := pr.NewHistogramVec(pr.HistogramOpts{
		Namespace: format(p.opts.Namespace),
		Name:      format(id),
		Help:      "histogram",
	}, fields)

	p.buf <- hv

	return &histogram{
		id: id,
		hv: hv,
		f:  p.opts.Fields,
	}
}

func (p *prometheus) String() string {
	return "prometheus"
}

func NewMetrics(opts ...metrics.Option) metrics.Metrics {
	return newMetrics(opts...)
}
