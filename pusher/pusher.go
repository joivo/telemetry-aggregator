package pusher

import (
	"log"
	"strings"

	"github.com/emanueljoivo/telemetry-aggregator/api"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type Pusher interface {
	PushMetric(o *Metric)
}

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

type PrometheusPusher struct{}

func (p PrometheusPusher) PushMetric(m *Metric) {
	var metricName = generateMetricName(m)

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        metricName,
		Help:        m.Help,
		ConstLabels: m.Metadata,
	})

	metric.Set(m.Value)

	log.Print("Pushing metric " + metricName + " to the Pushgateway.")

	if err := push.New(api.PushGatewayAddr, "probe_fogbow_stack").
		Collector(metric).
		Add(); err != nil {
		log.Println("Could not push completion time to Pushgateway: ", err)
	}
}

func generateMetricName(m *Metric) string {
	var resultName string
	const sep = "_"

	switch m.Name {
	case Reachability:
		resultName = ServiceLabel + sep + Reachability + sep + strings.ToLower(m.Metadata[ServiceLabel])
	case SuccessRate:
		resultName = ResourceLabel + sep + SuccessRate + sep + m.Metadata[ResourceLabel]
		m.Metadata[ServiceLabel] = "ras"
	case Latency:
		resultName = ResourceLabel + sep + Latency + sep + strings.ToLower(m.Metadata[ResourceLabel])
		m.Metadata[ServiceLabel] = "ras"
	case Availability:
		resultName = ResourceLabel + sep + Availability + sep + m.Metadata[ResourceLabel]
		m.Metadata[ServiceLabel] = "ras"
	default:
		resultName = ""
	}

	return resultName
}
