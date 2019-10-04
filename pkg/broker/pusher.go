package broker

import (
	"github.com/emanueljoivo/telemetry-aggregator/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
	"os"
)

type Pusher interface {
	PushMetric(o *models.Metric)
}

const PushGatewayAddr = "PUSHGATEWAY_ADDR"

type PrometheusPusher struct{}

func (p PrometheusPusher) PushMetric(m *models.Metric) {

	pushgatewayAddr, exists := os.LookupEnv(PushGatewayAddr)

	if exists {
		metric := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        m.Name,
			Help:        m.Help,
			ConstLabels: m.Metadata,
		})

		metric.Set(m.Value)

		log.Print("Pushing metric " + m.Name + " to the Pushgateway.")

		if err := push.New(pushgatewayAddr, "collect_fogbow_metric").
			Collector(metric).
			Add(); err != nil {
			log.Println("Could not push completion time to Pushgateway: ", err)
		}
	} else {
		log.Fatal("No push gateway address on the environment.")
	}
}
