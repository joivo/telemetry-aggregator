package pusher

import (
	"log"

	"github.com/emanueljoivo/telemetry-aggregator/api"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	ServiceReachability = "service_reachability"
	ServiceSuccessRate  = "service_success_rate"
	ServiceLatency      = "service_latency"
	ServiceAvailability = "service_availability"
)

type Pusher interface {
	Push(o *Metric)
}

type PrometheusPusher struct{}

func (p PrometheusPusher) Push(m *Metric) {
	go func() {
		metric := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: m.Name,
			Help: m.Help,
		})

		if err := push.New(api.PushGatewayAddr, "probe_by_observations").
			Collector(metric).
			Push(); err != nil {
			log.Println("Could not push completion time to Push Gateway: ", err)
		}
	}()
}

func generateLabels(m *Metric) {

	if m.Name == ServiceReachability {

	}

}
