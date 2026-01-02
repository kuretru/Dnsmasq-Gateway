package dnsmasq

import (
	"context"
	"time"
)

type Log struct {
	Context   context.Context
	Timestamp time.Time
	Message   string
}

type Point struct {
	Context   context.Context
	Timestamp time.Time
	Action    string
	Type      string
	Domain    string
	From      string
}
