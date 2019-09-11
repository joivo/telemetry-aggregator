package api

import "time"

// DefaultVersion of Current REST API
const (
	PushGatewayAddr    = "http://pushgateway:9091"

	DatabaseAddr       = "mongodb://db:27017"
	DatabaseName       = "aggregator"
	DatabaseCollection = "metrics"

	MetricEndpoint     = "/metric"
	VersionEndpoint    = "/version"

	DefaultVersion     = "v1.0.0"
	DefaultTimeout     = 10 * time.Second
)

type Version struct {
	Tag string `json:"tag"`
}