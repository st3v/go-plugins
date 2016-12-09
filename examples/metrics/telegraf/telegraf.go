package main

import (
	"fmt"
	"time"

	"github.com/micro/go-os/metrics"
	"github.com/micro/go-plugins/metrics/telegraf"
)

func main() {
	// create new metrics
	m := telegraf.NewMetrics(
		metrics.Namespace("io.micro"),
		metrics.WithFields(metrics.Fields{
			"region":  "eu-west-1",
			"service": "foo",
		}),
		metrics.Collectors(
			// telegraf/statsd address
			"127.0.0.1:8125",
		),
	)

	defer m.Close()

	c := m.Counter("example.counters.foo")
	g := m.Gauge("example.gauges.foo")
	h := m.Histogram("example.histograms.foo")

	for i := 0; i < 100; i++ {
		fmt.Println(time.Now().String(), "Sending metrics")
		c.Incr(uint64(i))
		g.Set(int64(i))
		h.Record(time.Now().Unix() * 1e3)
		time.Sleep(time.Millisecond * 10)
	}

	time.Sleep(time.Second * 5)
}
