/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package rest is the restful API router.
package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/metrics"
	"github.com/gin-gonic/gin"
)

// LogWriter defines the log writer.
type LogWriter interface {
	InfoWriter() io.Writer
	ErrorWriter() io.Writer
}

// Server defines the restful API server.
type Server struct {
	engine *gin.Engine
	// rg it contains the rest context, please use to realize some business logic.
	rg  *gin.RouterGroup
	ctx context.Context

	ip   string
	port int
	name string

	metrics *metrics.Monitor
}

// OptionFunc defines a function that can be used to modify the router.
type OptionFunc func(rg *gin.RouterGroup)

// WithPing with ping pong api.
func WithPing() OptionFunc {
	return func(rg *gin.RouterGroup) {
		rg.Any("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
}

// NewServer creates a new restful API server.
func NewServer(ctx context.Context, name, ip string, port int, logWriter LogWriter, apiOptFns ...OptionFunc) *Server {
	svr := &Server{
		ctx:    ctx,
		ip:     ip,
		port:   port,
		name:   name,
		engine: gin.New(),
	}

	// Recover from panic
	svr.engine.Use(gin.RecoveryWithWriter(logWriter.ErrorWriter()))

	// Set log middleware
	svr.engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    logWriter.InfoWriter(),
		Formatter: customLogFormatter,
		SkipPaths: []string{"/ping", "/healthz", "/metrics"},
	}))

	// Set metrics monitor.
	svr.metrics = metrics.NewMonitor(name).
		WithSlowTime(1 * time.Second).
		WithExcludePaths([]string{"/ping", "/healthz", "/metrics"}).
		RegisterMiddleware(svr.engine).
		Enable()

	svr.rg = svr.engine.Group("/")

	// Set authentication middleware.
	svr.rg.Use(MiddlewareContext())

	for _, fn := range apiOptFns {
		fn(svr.rg)
	}

	return svr
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
func (svr *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", svr.ip, svr.port)
	if err := svr.engine.Run(addr); err != nil {
		return err
	}

	return nil
}

// Name returns the router name.
func (svr *Server) Name() string {
	return svr.name
}

// IP returns the router ip.
func (svr *Server) IP() string {
	return svr.ip
}

// Port returns the router port.
func (svr *Server) Port() int {
	return svr.port
}
