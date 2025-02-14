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
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Enable enable metrics into prometheus handler.
func (monitor *Monitor) Enable() *Monitor {
	monitor.initMetrics()

	return monitor
}

// RegisterMiddleware is used to add monitor interceptor to gin router
// It can be called multiple times to intercept from multiple gin.IRoutes.
func (monitor *Monitor) RegisterMiddleware(r gin.IRoutes) *Monitor {
	r.Use(monitor.middleware)

	return monitor
}

// initMetrics used to init metrics.
func (monitor *Monitor) initMetrics() {
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
		Buckets:     monitor.durationMSBuckets,
	})
	_ = monitor.AddMetric(&metric{
		Type:        counter,
		Name:        monitor.metricKey.slowRequest,
		Description: fmt.Sprintf("the server handled slow requests counter, t=%dms.", monitor.slowTime.Milliseconds()),
		Labels:      []string{"uri", "method", "code"},
	})
}

// monitorMiddleware as gin monitor middleware.
func (monitor *Monitor) middleware(ctx *gin.Context) {
	// some paths should not be reported
	if slices.Contains(monitor.excludePaths, ctx.Request.URL.Path) {
		ctx.Next()

		return
	}
	startTime := time.Now()

	// execute normal process.
	ctx.Next()

	// after request
	monitor.metricHandle(&metricParam{
		request:               ctx.Request,
		requestPath:           ctx.FullPath(),
		responseStatusCode:    ctx.Writer.Status(),
		responseContentLength: int64(ctx.Writer.Size()),
		processDuration:       time.Since(startTime),
		clientIP:              ctx.ClientIP(),
	})
}

// HandleClientMetrics as a client side metrics handler.
func (monitor *Monitor) HandleClientMetrics(req *http.Request, resp *http.Response, subPath string, start time.Time) {
	if req == nil || resp == nil {
		return
	}

	monitor.metricHandle(&metricParam{
		request:               req,
		requestPath:           subPath,
		responseStatusCode:    resp.StatusCode,
		responseContentLength: resp.ContentLength,
		processDuration:       time.Since(start),
	})
}

type metricParam struct {
	request *http.Request

	// fullpath when handling server metrics.
	// subpath when handling client metrics.
	requestPath string

	responseStatusCode    int
	responseContentLength int64

	processDuration time.Duration

	// client side ip, empty when handling client metrics.
	clientIP string
}

// nolint:cyclop
func (monitor *Monitor) metricHandle(param *metricParam) {
	// set request total
	metric, err := monitor.getMetric(monitor.metricKey.requestTotal)
	if err == nil {
		_ = metric.Inc(nil)
	}

	// set uv
	if !monitor.bloomFilter.contains(param.clientIP) {
		monitor.bloomFilter.add(param.clientIP)
		if metric, err = monitor.getMetric(monitor.metricKey.requestUVTotal); err == nil {
			_ = metric.Inc(nil)
		}
	}

	// set uri request total
	if metric, err = monitor.getMetric(monitor.metricKey.uriRequestTotal); err == nil {
		_ = metric.Inc([]string{param.requestPath, param.request.Method, strconv.Itoa(param.responseStatusCode)})
	}

	// set request body size
	// since r.ContentLength can be negative (in some occasions) guard the operation
	if param.request.ContentLength >= 0 {
		if metric, err = monitor.getMetric(monitor.metricKey.requestBody); err == nil {
			_ = metric.Add(nil, float64(param.request.ContentLength))
		}
	}

	// set slow request
	if param.processDuration >= monitor.slowTime {
		if metric, err = monitor.getMetric(monitor.metricKey.slowRequest); err == nil {
			_ = metric.Inc([]string{param.requestPath, param.request.Method, strconv.Itoa(param.responseStatusCode)})
		}
	}

	// set request duration
	if metric, err = monitor.getMetric(monitor.metricKey.requestDuration); err == nil {
		_ = metric.Observe([]string{param.requestPath}, float64(param.processDuration.Milliseconds()))
	}

	// set response size
	if param.responseContentLength > 0 {
		if metric, err = monitor.getMetric(monitor.metricKey.responseBody); err == nil {
			_ = metric.Add(nil, float64(param.responseContentLength))
		}
	}
}
