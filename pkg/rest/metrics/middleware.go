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
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Use set gin metrics middleware.
func (monitor *Monitor) Use(r gin.IRoutes) {
	monitor.initGinMetrics()

	r.Use(monitor.middleware)
	r.GET(monitor.metricPath, gin.WrapH(promhttp.Handler()))
}

// UseWithoutExposingEndpoint is used to add monitor interceptor to gin router
// It can be called multiple times to intercept from multiple gin.IRoutes
// http path is not set, to do that use Expose function.
func (monitor *Monitor) UseWithoutExposingEndpoint(r gin.IRoutes) {
	monitor.initGinMetrics()
	r.Use(monitor.middleware)
}

// Expose adds metric path to a given router.
// The router can be different with the one passed to UseWithoutExposingEndpoint.
// This allows to expose metrics on different port.
func (monitor *Monitor) Expose(r gin.IRoutes) {
	r.GET(monitor.metricPath, gin.WrapH(promhttp.Handler()))
}

// initGinMetrics used to init gin metrics.
func (monitor *Monitor) initGinMetrics() {
	monitor.bloomFilter = newBloomFilter()

	_ = monitor.AddMetric(&metric{
		Type:        counter,
		Name:        monitor.metricKey.requestTotal,
		Description: "all the server received request num.",
		Labels:      nil,
	})
	_ = monitor.AddMetric(&metric{
		Type:        counter,
		Name:        monitor.metricKey.requestUVTotal,
		Description: "all the server received ip num.",
		Labels:      nil,
	})
	_ = monitor.AddMetric(&metric{
		Type:        counter,
		Name:        monitor.metricKey.uriRequestTotal,
		Description: "all the server received request num with every uri.",
		Labels:      []string{"uri", "method", "code"},
	})
	_ = monitor.AddMetric(&metric{
		Type:        counter,
		Name:        monitor.metricKey.requestBody,
		Description: "the server received request body size, unit byte",
		Labels:      nil,
	})
	_ = monitor.AddMetric(&metric{
		Type:        counter,
		Name:        monitor.metricKey.responseBody,
		Description: "the server send response body size, unit byte",
		Labels:      nil,
	})
	_ = monitor.AddMetric(&metric{
		Type:        histogram,
		Name:        monitor.metricKey.requestDuration,
		Description: "the time server took to handle the request.",
		Labels:      []string{"uri"},
		Buckets:     monitor.reqDuration,
	})
	_ = monitor.AddMetric(&metric{
		Type:        counter,
		Name:        monitor.metricKey.slowRequest,
		Description: fmt.Sprintf("the server handled slow requests counter, t=%d.", monitor.slowTime),
		Labels:      []string{"uri", "method", "code"},
	})
}

// monitorMiddleware as gin monitor middleware.
func (monitor *Monitor) middleware(ctx *gin.Context) {
	// some paths should not be reported
	if ctx.Request.URL.Path == monitor.metricPath ||
		slices.Contains(monitor.excludePaths, ctx.Request.URL.Path) {

		ctx.Next()

		return
	}
	startTime := time.Now()

	// execute normal process.
	ctx.Next()

	// after request
	monitor.ginMetricHandle(ctx, startTime)
}

func (monitor *Monitor) ginMetricHandle(ctx *gin.Context, start time.Time) {
	request := ctx.Request
	writer := ctx.Writer

	// set request total
	metric, err := monitor.getMetric(monitor.metricKey.requestTotal)
	if err == nil {
		_ = metric.Inc(nil)
	}

	// set uv
	if clientIP := ctx.ClientIP(); !monitor.bloomFilter.contains(clientIP) {
		monitor.bloomFilter.add(clientIP)
		if metric, err = monitor.getMetric(monitor.metricKey.requestUVTotal); err == nil {
			_ = metric.Inc(nil)
		}
	}

	// set uri request total
	if metric, err = monitor.getMetric(monitor.metricKey.uriRequestTotal); err == nil {
		_ = metric.Inc([]string{ctx.FullPath(), request.Method, strconv.Itoa(writer.Status())})
	}

	// set request body size
	// since r.ContentLength can be negative (in some occasions) guard the operation
	if request.ContentLength >= 0 {
		if metric, err = monitor.getMetric(monitor.metricKey.requestBody); err == nil {
			_ = metric.Add(nil, float64(request.ContentLength))
		}
	}

	// set slow request
	latency := time.Since(start)
	if int32(latency.Seconds()) > monitor.slowTime {
		if metric, err = monitor.getMetric(monitor.metricKey.slowRequest); err == nil {
			_ = metric.Inc([]string{ctx.FullPath(), request.Method, strconv.Itoa(writer.Status())})
		}
	}

	// set request duration
	if metric, err = monitor.getMetric(monitor.metricKey.requestDuration); err == nil {
		_ = metric.Observe([]string{ctx.FullPath()}, latency.Seconds())
	}

	// set response size
	if writer.Size() > 0 {
		if metric, err = monitor.getMetric(monitor.metricKey.responseBody); err == nil {
			_ = metric.Add(nil, float64(writer.Size()))
		}
	}
}
