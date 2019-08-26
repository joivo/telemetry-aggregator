package main

import (
	"log"
)

type Metric struct {
	Description string `json:"description"`
	Measurement int64  `json:"measurement"`
}

type Observation struct {
	Label     string   `json:"label"`
	Timestamp int64    `json:"timestamp"`
	Values    []Metric `json:"values"`
}

func main() {
	log.Println(Observation{})
}
