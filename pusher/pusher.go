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
	ResourceAvailability = "resource_availability"
)

const (
	ServiceLabel  = "service"
	ResourceLabel = "resource"
)

type Pusher interface {
	Push(o *Metric)
}

type PrometheusPusher struct{}

func (p PrometheusPusher) Push(m *Metric) {
	go func() {
		var labels = generateLabels(m)

		metric := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: m.Name,
			Help: m.Help,
			ConstLabels: labels,
		})

		if err := push.New(api.PushGatewayAddr, "probing_structure").
			Collector(metric).
			Push(); err != nil {
			log.Println("Could not push completion time to Push Gateway: ", err)
		}
	}()
}

func generateLabels(m *Metric) map[string]string {

	var labels = make(map[string]string)

	if m.Name == ServiceReachability {
		labels[ServiceLabel] = "don't supported yet"
	} else {
		labels[ServiceLabel] = "ras"
		labels[ResourceLabel] = "don't supported yet"
	}

	return labels
}
