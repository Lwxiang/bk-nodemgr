/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package service provides saas service.
package service

import (
	"context"
	"io"
	"runtime"

	"git.woa.com/bk-gse/bk-nodeman/internal/saas/options"
	"git.woa.com/bk-gse/bk-nodeman/internal/saas/router/metrics"
	"git.woa.com/bk-gse/bk-nodeman/internal/saas/router/web"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/gopool"
	"github.com/gin-gonic/gin"
)

// Service defines a server that provides saas services.
// It provides a website for user to operate with nodeman.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.SaasService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx context.Context
	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*rest.Server

	// Note: Capability is initialized in the Start() and could not be used in other package.
	// Capability is the capability of the service.
	Capability *options.Capability
}

const (
	// RouterNameHTTPServer defines the name of http server router.
	RouterNameHTTPServer = "http-server"
)

// NewService creates a new saas service.
func NewService(conf *config.SaasService) *Service {
	svc := &Service{
		conf: conf,
	}
	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	httpServer := rest.NewServer(svc.ctx, RouterNameHTTPServer, conf.HTTPServer.BindIP, conf.HTTPServer.Port,
		loggerWriter{},
		rest.WithPing(),
		withWeb(svc.Capability),
		withMetrics(svc.Capability),
	)
	svc.servers = append(svc.servers, httpServer)

	return svc
}

// loggerWriter implements rest.LoggerWriter.
type loggerWriter struct{}

func (l loggerWriter) InfoWriter() io.Writer {
	return blog.WriterInfo{}
}

func (l loggerWriter) ErrorWriter() io.Writer {
	return blog.WriterError{}
}

// withWeb load web page handler.
func withWeb(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		web.Load(rg, capability)
	}
}

// withMetrics load metrics handler.
func withMetrics(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		metrics.Load(rg, capability)
	}
}

// Start starts the saas service.
func (svc *Service) Start(ctx context.Context) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logConfig := blog.NewLogConfig()
	logConfig.LogDir = svc.conf.Log.Dir
	logConfig.LogMaxSizeMB = svc.conf.Log.MaxSizeMB
	logConfig.LogMaxNum = svc.conf.Log.MaxNum
	logConfig.Level = svc.conf.Log.Level
	logConfig.ToStdErr = svc.conf.Log.ToStdErr
	logConfig.AlsoToStdErr = svc.conf.Log.AlsoToStdErr
	blog.InitLogs(logConfig)

	svc.ctx, svc.cancelFunc = context.WithCancel(ctx)

	// start servers
	gp := gopool.NewPool()
	for _, server := range svc.servers {
		// server start will block until router stop, so we need to run it in a goroutine.
		fn := func() error {
			if err := server.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
		blog.Infof("started server. name(%s), ip(%s), port(%d)", server.Name(), server.IP(), server.Port())
	}

	// wait until all routers stopped or application error.
	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to start servers, err: %v", err)
		return err
	}

	return nil
}
