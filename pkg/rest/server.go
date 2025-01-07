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
	s := &Server{
		ctx:    ctx,
		ip:     ip,
		port:   port,
		name:   name,
		engine: gin.New(),
	}

	// Recover from panic
	s.engine.Use(gin.RecoveryWithWriter(logWriter.ErrorWriter()))

	// Set log middleware
	s.engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    logWriter.InfoWriter(),
		Formatter: customLogFormatter}))

	s.engine.Use()

	s.rg = s.engine.Group("/")

	// Set authentication middleware.
	s.rg.Use(MiddlewareContext())

	for _, fn := range apiOptFns {
		fn(s.rg)
	}

	return s
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
func (r *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", r.ip, r.port)
	if err := r.engine.Run(addr); err != nil {
		return err
	}

	return nil
}

// Name returns the router name.
func (r *Server) Name() string {
	return r.name
}

// IP returns the router ip.
func (r *Server) IP() string {
	return r.ip
}

// Port returns the router port.
func (r *Server) Port() int {
	return r.port
}
