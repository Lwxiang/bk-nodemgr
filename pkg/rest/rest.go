/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rest xxx
package rest

import (
	"context"
	"net/http"

	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/errf"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/tracing"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// Response standard response
type Response struct {
	Result    bool        `json:"result"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// HandlerFunc xxx
type HandlerFunc func(*Context) (interface{}, error)

// StreamHandlerFunc xxx
type StreamHandlerFunc func(*Context)

// AbortWithBadRequestError 请求失败
func AbortWithBadRequestError(c *Context, err error) {
	result := Response{Code: errf.InvalidParameter, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusBadRequest, result)
}

// AbortWithUnauthorizedError 未登入
func AbortWithUnauthorizedError(c *Context, err error) {
	result := Response{Code: errf.DoAuthorizeFailed, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusUnauthorized, result)
}

// AbortWithWithForbiddenError no permission
func AbortWithWithForbiddenError(c *Context, err error) {
	result := Response{Code: errf.PermissionDenied, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusForbidden, result)
}

// AbortWithJSONError xxx
func AbortWithJSONError(c *Context, err error) {
	// TODO: support error code
	result := Response{Code: errf.Aborted, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusOK, result)
}

// APIResponse 正常返回
func APIResponse(c *Context, data interface{}) {
	result := Response{Code: 0, Message: "OK", RequestId: c.RequestId, Data: data}
	c.JSON(http.StatusOK, result)
}

// InitRestContext ...
func InitRestContext(c *gin.Context) *Context {
	requestId := requestid.Get(c)

	restContext := &Context{
		Context:   c,
		RequestId: requestId,
	}
	c.Set("rest_context", restContext)

	tracing.SetRequestIDValue(c.Request, requestId)
	ctx := context.WithValue(c.Request.Context(), tracing.RequestIDHeaderKey, requestId)

	restContext.Request = restContext.Request.WithContext(ctx)
	return restContext
}

// GetRestContext only when user has authenticated, otherwise, return ErrorUnauthorized
func GetRestContext(c *gin.Context) (*Context, error) {
	ctxObj, ok := c.Get("rest_context")
	if !ok {
		return nil, errf.ErrorUnauthorized
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, errf.ErrorUnauthorized
	}

	return restContext, nil
}

// RestHandlerFunc rest handler
func RestHandlerFunc(handler HandlerFunc) gin.HandlerFunc { // nolint
	return func(c *gin.Context) {
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		result, err := handler(restContext)
		if err != nil {
			AbortWithJSONError(restContext, err)
			return
		}

		APIResponse(restContext, result)
	}
}

// STDRestHandlerFunc std rest handler
func STDRestHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		result, err := handler(restContext)
		if err != nil {
			AbortWithBadRequestError(restContext, err)
			return
		}

		APIResponse(restContext, result)
	}
}

// StreamHandler 流式 Handler
func StreamHandler(handler StreamHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		handler(restContext)
	}
}
