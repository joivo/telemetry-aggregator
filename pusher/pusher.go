package pusher

import (
	"log"

	"github.com/emanueljoivo/telemetry-aggregator/api"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	Reachability = "reachability"
	SuccessRate  = "success_rate"
	Latency      = "latency"
	Availability = "availability"
)

const (
	ServiceLabel  = "service"
	ResourceLabel = "resource"
)

type Pusher interface {
	PushMetric(o *Metric)
}

type PrometheusPusher struct{}

func (p PrometheusPusher) PushMetric(m *Metric) {
	go func() {

		metric := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: m.Name,
			Help: m.Help,
		})

		log.Println(metric)

		if err := push.New(api.PushGatewayAddr, "probing_structure").
			Collector(metric).
			Push(); err != nil {
			log.Println("Could not push completion time to Push Gateway: ", err)
		}
	}()
}
