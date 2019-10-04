package broker

import "github.com/emanueljoivo/telemetry-aggregator/pkg/models"

var Metrics = make(map[int64]models.Metric)

// Rests a new metric to the pool
func Place(m *models.Metric) {
	Metrics[m.Timestamp] = *m
}

// Empties the pool removing all the current metrics that has on it
func Deflate() map[int64]models.Metric {
	return Metrics
}