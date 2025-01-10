/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package metrics provides metrics for gin.
package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// Metric defines a metric object. Users can use it to save
// metric data. Every metric should be globally unique by name.
type metric struct {
	Type        metricType
	Name        string
	Description string
	Labels      []string
	Buckets     []float64
	Objectives  map[float64]float64

	vec prometheus.Collector
}

func generateErrorDoesNotExist(name string) error {
	return fmt.Errorf("metric does not exist. name(%s)", name)
}

func generateErrorTypeNotSupport(t metricType, name string) error {
	return fmt.Errorf("metric not support type. name(%s), type(%d)", name, t)
}

func generateErrorNotSpecificType(t metricType, name string) error {
	switch t {
	case none:
		return fmt.Errorf("metric is not none type. name(%s)", name)
	case gauge:
		return fmt.Errorf("metric is not gauge type. name(%s)", name)
	case counter:
		return fmt.Errorf("metric is not counter type. name(%s)", name)
	case histogram:
		return fmt.Errorf("metric is not histogram type. name(%s)", name)
	case summary:
		return fmt.Errorf("metric is not summary type. name(%s)", name)
	default:
		return fmt.Errorf("metric type is not specified. name(%s)", name)
	}
}

// SetGaugeValue set data for Gauge type Metric.
func (mc *metric) SetGaugeValue(labelValues []string, value float64) error {
	switch mc.Type {
	case none:
		return generateErrorDoesNotExist(mc.Name)

	case gauge:
		vec, ok := mc.vec.(*prometheus.GaugeVec)
		if !ok {
			return generateErrorNotSpecificType(gauge, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Set(value)

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}

// Inc increases value for Counter/Gauge type metric, increments
// the counter by 1.
func (mc *metric) Inc(labelValues []string) error {
	switch mc.Type {
	case none:
		return generateErrorDoesNotExist(mc.Name)

	case gauge:
		vec, ok := mc.vec.(*prometheus.GaugeVec)
		if !ok {
			return generateErrorNotSpecificType(gauge, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Inc()

	case counter:
		vec, ok := mc.vec.(*prometheus.CounterVec)
		if !ok {
			return generateErrorNotSpecificType(counter, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Inc()

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}

// Add adds the given value to the Metric object. Only
// for Counter/Gauge type metric.
func (mc *metric) Add(labelValues []string, value float64) error {
	switch mc.Type {
	case none:
		return generateErrorDoesNotExist(mc.Name)

	case gauge:
		vec, ok := mc.vec.(*prometheus.GaugeVec)
		if !ok {
			return generateErrorNotSpecificType(gauge, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Add(value)

	case counter:
		vec, ok := mc.vec.(*prometheus.CounterVec)
		if !ok {
			return generateErrorNotSpecificType(counter, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Add(value)

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}

// Observe is used by Histogram and Summary type metric to
// add observations.
func (mc *metric) Observe(labelValues []string, value float64) error {
	switch mc.Type {
	case none:
		return generateErrorDoesNotExist(mc.Name)

	case histogram:
		vec, ok := mc.vec.(*prometheus.HistogramVec)
		if !ok {
			return generateErrorNotSpecificType(histogram, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Observe(value)

	case summary:
		vec, ok := mc.vec.(*prometheus.SummaryVec)
		if !ok {
			return generateErrorNotSpecificType(summary, mc.Name)
		}
		vec.WithLabelValues(labelValues...).Observe(value)

	default:
		return generateErrorTypeNotSupport(mc.Type, mc.Name)
	}

	return nil
}
