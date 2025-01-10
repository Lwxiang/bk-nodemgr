/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package metrics

import (
	"errors"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type metricType int

const (
	none metricType = iota
	counter
	gauge
	histogram
	summary

	defaultMetricPath = "/metrics"
	defaultSlowTime   = int32(5)
)

func defaultExcludePaths() []string {
	return []string{}
}

func defaultRequestDuration() []float64 {
	return []float64{0.1, 0.3, 1.2, 5, 10}
}

type metricKey struct {
	requestTotal    string
	requestUVTotal  string
	uriRequestTotal string
	requestBody     string
	responseBody    string
	requestDuration string
	slowRequest     string
}

// Monitor is an object that uses to set gin server monitor.
type Monitor struct {
	slowTime     int32
	metricPath   string
	excludePaths []string
	reqDuration  []float64
	metrics      map[string]*metric
	bloomFilter  *bloomFilter
	typeHandler  map[metricType]func(metric *metric) error
	metricKey    *metricKey
}

// nolint:gochecknoglobals
var (
	once    sync.Once
	monitor *Monitor
)

// GetMonitor used to get global Monitor object,
// this function returns a singleton object.
func GetMonitor() *Monitor {
	once.Do(func() {
		monitor = &Monitor{
			metricPath:   defaultMetricPath,
			slowTime:     defaultSlowTime,
			excludePaths: defaultExcludePaths(),
			reqDuration:  defaultRequestDuration(),
			metrics:      make(map[string]*metric),
			bloomFilter:  newBloomFilter(),
			typeHandler: map[metricType]func(metric *metric) error{
				counter:   counterHandler,
				gauge:     gaugeHandler,
				histogram: histogramHandler,
				summary:   summaryHandler,
			},
			metricKey: &metricKey{
				requestTotal:    "metric_request_total",
				requestUVTotal:  "metric_request_uv_total",
				uriRequestTotal: "metric_uri_request_total",
				requestBody:     "metric_request_body",
				responseBody:    "metric_response_body",
				requestDuration: "metric_request_duration",
				slowRequest:     "metric_slow_request",
			},
		}
	})

	return monitor
}

// GetMetric used to get metric object by metric_name.
func (monitor *Monitor) getMetric(name string) (*metric, error) {
	if metric, ok := monitor.metrics[name]; ok {
		return metric, nil
	}

	return nil, fmt.Errorf("metric not found. name(%s)", name)
}

// SetMetricPath set metricPath property. metricPath is used for Prometheus
// to get gin server monitoring data.
func (monitor *Monitor) SetMetricPath(path string) {
	monitor.metricPath = path
}

// SetExcludePaths set exclude paths which should not be reported (e.g. /ping /healthz...)
func (monitor *Monitor) SetExcludePaths(paths []string) {
	monitor.excludePaths = paths
}

// SetSlowTime set slowTime property. slowTime is used to determine whether
// the request is slow. For "gin_slow_request_total" metric.
func (monitor *Monitor) SetSlowTime(slowTime int32) {
	monitor.slowTime = slowTime
}

// SetDuration set reqDuration property. reqDuration is used to ginRequestDuration
// metric buckets.
func (monitor *Monitor) SetDuration(duration []float64) {
	monitor.reqDuration = duration
}

// SetMetricPrefix set metric prefix.
func (monitor *Monitor) SetMetricPrefix(prefix string) {
	monitor.metricKey.requestTotal = prefix + monitor.metricKey.requestTotal
	monitor.metricKey.requestUVTotal = prefix + monitor.metricKey.requestUVTotal
	monitor.metricKey.uriRequestTotal = prefix + monitor.metricKey.uriRequestTotal
	monitor.metricKey.requestBody = prefix + monitor.metricKey.requestBody
	monitor.metricKey.responseBody = prefix + monitor.metricKey.responseBody
	monitor.metricKey.requestDuration = prefix + monitor.metricKey.requestDuration
	monitor.metricKey.slowRequest = prefix + monitor.metricKey.slowRequest
}

// SetMetricSuffix set metric suffix.
func (monitor *Monitor) SetMetricSuffix(suffix string) {
	monitor.metricKey.requestTotal += suffix
	monitor.metricKey.requestUVTotal += suffix
	monitor.metricKey.uriRequestTotal += suffix
	monitor.metricKey.requestBody += suffix
	monitor.metricKey.responseBody += suffix
	monitor.metricKey.requestDuration += suffix
	monitor.metricKey.slowRequest += suffix
}

// AddMetric add custom monitor metric.
func (monitor *Monitor) AddMetric(metric *metric) error {
	if _, ok := monitor.metrics[metric.Name]; ok {
		return fmt.Errorf("metric already exists. name(%s)", metric.Name)
	}

	if metric.Name == "" {
		return errors.New("metric name cannot be empty")
	}
	if f, ok := monitor.typeHandler[metric.Type]; ok {
		if err := f(metric); err == nil {
			prometheus.MustRegister(metric.vec)
			monitor.metrics[metric.Name] = metric

			return nil
		}
	}

	return generateErrorTypeNotSupport(metric.Type, metric.Name)
}

func counterHandler(metric *metric) error {
	metric.vec = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: metric.Name, Help: metric.Description},
		metric.Labels,
	)

	return nil
}

func gaugeHandler(metric *metric) error {
	metric.vec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: metric.Name, Help: metric.Description},
		metric.Labels,
	)

	return nil
}

func histogramHandler(metric *metric) error {
	if len(metric.Buckets) == 0 {
		return fmt.Errorf("histogram type must specify buckets. name(%s)", metric.Name)
	}

	metric.vec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    metric.Name,
			Help:    metric.Description,
			Buckets: metric.Buckets,
		},
		metric.Labels,
	)

	return nil
}

func summaryHandler(metric *metric) error {
	if len(metric.Objectives) == 0 {
		return fmt.Errorf("summary type must specify objectives. name(%s)", metric.Name)
	}

	metric.vec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       metric.Name,
			Help:       metric.Description,
			Objectives: metric.Objectives,
		},
		metric.Labels,
	)

	return nil
}
