package main

import "time"

type Target struct {
	Name   string
	Host   string
	Kind   string
	Mode   string
	Custom bool
}

type PingResult struct {
	Timestamp time.Time
	Target    Target
	Success   bool
	Latency   float64
	Message   string
}

type Sample struct {
	Time    time.Time
	Latency float64
	Success bool
}

type Outage struct {
	Start    time.Time
	End      time.Time
	Category string
	Details  string
}
