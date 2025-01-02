/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package router is the restful API router.
package router

import (
	"context"
	"fmt"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/internal/options"
	"git.woa.com/bk-gse/bk-nodeman/internal/router/api-v3"
	"git.woa.com/bk-gse/bk-nodeman/internal/router/basic"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	middleware "git.woa.com/bk-gse/bk-nodeman/pkg/rest/middlerware"
	"github.com/gin-gonic/gin"
)

// Router defines the router.
type Router struct {
	engine *gin.Engine
	// rg it contains the rest context, please use to realize some business logic.
	rg   *gin.RouterGroup
	cap  *options.Capability
	ctx  context.Context
	ip   string
	port int
	name string
}

// OptionFunc defines a function that can be used to modify the router.
type OptionFunc func(*Router)

// WithApiV3 adds the api v3 router.
func WithApiV3() OptionFunc {
	return func(r *Router) {
		apiv3.Load(r.rg, r.cap)
	}
}

// WithBasic adds the basic router.
func WithBasic() OptionFunc {
	return func(r *Router) {
		basic.Load(r.rg, r.cap)
	}
}

// NewRouter creates a new router.
func NewRouter(ctx context.Context, name string, ip string, port int, capability *options.Capability,
	optFns ...OptionFunc) *Router {

	r := &Router{
		ctx:    ctx,
		name:   name,
		ip:     ip,
		port:   port,
		engine: gin.New(),
		cap:    capability,
	}

	// Recover from panic
	r.engine.Use(gin.RecoveryWithWriter(blog.GlogWriter{}))

	// Set log middleware
	r.engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    blog.GlogWriter{},
		Formatter: customLogFormatter}))

	r.engine.Use()

	r.rg = r.engine.Group("/")
	r.rg.Any("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Set authentication middleware
	r.rg.Use(middleware.InitRestContext())

	for _, optFn := range optFns {
		optFn(r)
	}

	return r
}

// customLogFormatter is a custom log formatter.
func customLogFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	return fmt.Sprintf("[GIN Requst] |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

// Start starts the router.
func (r *Router) Start() error {
	addr := fmt.Sprintf("%s:%d", r.ip, r.port)
	if err := r.engine.Run(addr); err != nil {
		return err
	}

	return nil
}

// Name returns the router name.
func (r *Router) Name() string {
	return r.name
}

// IP returns the router ip.
func (r *Router) IP() string {
	return r.ip
}

// Port returns the router port.
func (r *Router) Port() int {
	return r.port
}
