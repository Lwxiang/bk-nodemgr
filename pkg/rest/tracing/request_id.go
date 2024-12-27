/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tracing request id
package tracing

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	RequestIDHeaderKey = "X-Bkapi-Request-Id"
)

// RequestIdGenerator :
func RequestIdGenerator() string {
	uid := uuid.New().String()
	requestId := strings.ReplaceAll(uid, "-", "")
	return requestId
}

// SetRequestIDValue :
func SetRequestIDValue(req *http.Request, id string) {
	req.Header.Set(RequestIDHeaderKey, id)
}

// RequestIDValue :
func RequestIDValue(req *http.Request, autoGen bool) string {
	id := req.Header.Get(RequestIDHeaderKey)
	if id == "" && autoGen {
		id = RequestIdGenerator()
	}
	return id
}
