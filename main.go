package main

import (
	"fmt"
	"time"
)

type Observation struct {
	Timestamp int64
	Value string
}

func main() {
	fmt.Println(Observation{
		time.Now().Unix(),
		"10 successful requests, 0 failed requests",
	})
}