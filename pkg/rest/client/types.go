package client

import (
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/discovery"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ToleranceLatencyTimeDefault = 500 * time.Millisecond
)

// Capability http request limit.
type Capability struct {
	// Client http client.
	Client HTTPClient

	// Discover get request address.
	Discover discovery.Interface

	// the max tolerance api request latency time, if exceeded this time, then
	// this request will be logged and warned.
	ToleranceLatencyTime time.Duration

	// MetricOpts metric option.
	MetricOpts MetricOption

	// Logger logger
	Logger Logger
}

// MetricOption metrics options.
type MetricOption struct {
	// prometheus metric register
	Register prometheus.Registerer
	// if not set, use default buckets value
	DurationBuckets []float64
}

// Logger is the logger interface.
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugw(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infow(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnw(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorw(args ...interface{})
}
