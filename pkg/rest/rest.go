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
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/header"
	"github.com/gin-gonic/gin"
)

// Response standard response.
type Response struct {
	Result    bool        `json:"result"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// HandlerFunc defines the router handler.
type HandlerFunc func(*Context) (interface{}, error)

// StreamHandlerFunc defines the stream handler.
type StreamHandlerFunc func(*Context)

// AbortWithBadRequestError provides handler process failing response.
func AbortWithBadRequestError(c *Context, err error) {
	result := Response{Code: errf.InvalidParameter, Message: err.Error(), RequestID: c.RequestID}
	c.AbortWithStatusJSON(http.StatusBadRequest, result)
}

// AbortWithUnauthorizedError provides auth check failing response.
func AbortWithUnauthorizedError(c *Context, err error) {
	result := Response{Code: errf.DoAuthorizeFailed, Message: err.Error(), RequestID: c.RequestID}
	c.AbortWithStatusJSON(http.StatusUnauthorized, result)
}

// AbortWithWithForbiddenError provides permission denied response.
func AbortWithWithForbiddenError(c *Context, err error) {
	result := Response{Code: errf.PermissionDenied, Message: err.Error(), RequestID: c.RequestID}
	c.AbortWithStatusJSON(http.StatusForbidden, result)
}

// AbortWithJSONError provides handler process failing response.
func AbortWithJSONError(ctx *Context, err error) {
	// TODO: support error code
	result := Response{Code: errf.Aborted, Result: false, Message: err.Error(), RequestID: ctx.RequestID}
	ctx.AbortWithStatusJSON(http.StatusOK, result)
}

// APIResponse provides handler process successfully and make a normal response.
func APIResponse(ctx *Context, data interface{}) {
	result := Response{Code: 0, Result: true, Message: "OK", RequestID: ctx.RequestID, Data: data}
	ctx.JSON(http.StatusOK, result)
}

// restContextKey was used to store the restContext in gin.Context.
const restContextKey = "rest_context"

// InitRestContext initializes a new rest context.
func InitRestContext(pCtx *gin.Context) *Context {
	restContext := &Context{
		Context:   pCtx,
		RequestID: header.RequestIDValue(pCtx.Request, true),
		Username:  pCtx.GetHeader(header.UserKey),
	}

	pCtx.Set(restContextKey, restContext)

	// note: for thread safety you need to reset it here.
	ctx := context.WithValue(pCtx.Request.Context(), header.RIDKey, restContext.RequestID)
	restContext.Request = restContext.Request.WithContext(ctx)

	return restContext
}

// GetRestContext only when user has authenticated, otherwise, return ErrorUnauthorized.
func GetRestContext(c *gin.Context) (*Context, error) {
	ctxObj, ok := c.Get(restContextKey)
	if !ok {
		return nil, errf.ErrorUnauthorized
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, errf.ErrorUnauthorized
	}

	return restContext, nil
}

// RestHandlerFunc rest handler.
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

// STDRestHandlerFunc std rest handler.
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

// StreamHandler stream handler.
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
