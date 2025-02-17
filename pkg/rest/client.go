/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package rest

import (
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/rest/metrics"
)

const (
	// maxRetryCycleDefault max retry cycle.
	maxRetryCycleDefault = 3

	// toleranceLatencyTimeDefault tolerance latency time.
	toleranceLatencyTimeDefault = 2 * time.Second
)

// ClientInterface http client interface.
type ClientInterface interface {
	Verb(verb VerbType) *Request
	Post() *Request
	Put() *Request
	Get() *Request
	Delete() *Request
	Patch() *Request
	Head() *Request
}

// NewClient get rest client.
func NewClient(capability *client.Capability, baseURL string) (ClientInterface, error) {
	if baseURL != "/" {
		baseURL = strings.Trim(baseURL, "/")
		baseURL = "/" + baseURL + "/"
	}

	if capability.ToleranceLatencyTime <= 0 {
		// set default tolerance latency time
		capability.ToleranceLatencyTime = toleranceLatencyTimeDefault
	}

	client := &Client{
		baseURL:       baseURL,
		capability:    capability,
		maxRetryCycle: maxRetryCycleDefault,
		metrics: metrics.NewMonitor(capability.Name).
			WithDurationMSBuckets(capability.MetricOpts.DurationMSBuckets).
			WithSlowTime(capability.ToleranceLatencyTime).
			Enable(),
	}

	return client, nil
}

// Client http client.
type Client struct {
	// base url.
	baseURL string

	// client capability.
	capability *client.Capability

	// client metrics monitor.
	metrics *metrics.Monitor

	// exclusionURL
	exclusionURL []string

	// maxRetryCycle
	maxRetryCycle int
}

// Verb get request.
func (r *Client) Verb(verb VerbType) *Request {
	return &Request{
		client:     r,
		verb:       verb,
		baseURL:    r.baseURL,
		capability: r.capability,
	}
}

// Post method.
func (r *Client) Post() *Request {
	return r.Verb(VerbTypePOST)
}

// Put method.
func (r *Client) Put() *Request {
	return r.Verb(VerbTypePUT)
}

// Get method.
func (r *Client) Get() *Request {
	return r.Verb(VerbTypeGET)
}

// Delete delete method.
func (r *Client) Delete() *Request {
	return r.Verb(VerbTypeDELETE)
}

// Patch method.
func (r *Client) Patch() *Request {
	return r.Verb(VerbTypePATCH)
}

// Head method.
func (r *Client) Head() *Request {
	return r.Verb(VerbTypeHEAD)
}
