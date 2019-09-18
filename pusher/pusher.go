package pusher

import (
	"log"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type Pusher interface {
	PushMetric(o *Metric)
}

const (
	PushGatewayAddr = "PUSHGATEWAY_ADDR"
	Reachability = "reachability"
	SuccessRate  = "success_rate"
	Latency      = "latency"
	Availability = "availability"
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

	pushgatewayAddr, exists := os.LookupEnv(PushGatewayAddr)

	if exists {
		if err := push.New(pushgatewayAddr, "collect_fogbow_metric").
			Collector(metric).
			Add(); err != nil {
			log.Println("Could not push completion time to Pushgateway: ", err)
		}
	} else {
		log.Fatal("No push gateway address on the environment.")
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
