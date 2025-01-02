/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package backend provides backend service.
package backend

import (
	"context"
	"runtime"

	"git.woa.com/bk-gse/bk-nodeman/internal/manager"
	"git.woa.com/bk-gse/bk-nodeman/internal/options"
	"git.woa.com/bk-gse/bk-nodeman/internal/router"
	topoStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage/topo"
	"git.woa.com/bk-gse/bk-nodeman/pkg/apigw"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/cmdb"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/gopool"
)

// Service defines a server that provides backend services.
// It manages the configuration, lifecycle, and various capabilities (e.g., cmdb, topo storage).
// The service's capabilities are accessed through its 'cap' field, while 'router' is used to route requests.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.BackendService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx context.Context
	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// Note: the following fields are initialized in the Start() and could not be used in other package.
	// manager workflow management.
	manager *manager.Manager
	// topoStorage bk nodeman topo storage
	topoStorage *topoStorage.DefaultStorage
	// cmdbHandler cmdb handler
	cmdbHandler cmdb.Handler
	// apigwCli apigw client
	apigwCli apigw.Client

	// router is the entry point of the service, routing requests to different capabilities.
	routers []*router.Router

	// Note: Capability is initialized in the Start() and could not be used in other package.
	// Capability is the capability of the service.
	Capability *options.Capability
}

const (
	// TopoStorageDatabase bk node manager database name.
	TopoStorageDatabase = "bknodeman_topo"
	// TopoStorageBusinessCollection bk node manager business collection name.
	TopoStorageBusinessCollection = "business"
	// TopoStorageHostCollection bk node manager host collection name.
	TopoStorageHostCollection = "host"
)

// NewService creates a new backend service.
func NewService(conf *config.BackendService) *Service {
	svc := &Service{
		conf: conf,
	}
	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	svc.apigwCli = apigw.NewClient(&apigw.Config{
		BKAppCode:        conf.APIGateway.AppCode,
		BKAppSecret:      conf.APIGateway.AppSecret,
		PlatformUsername: conf.APIGateway.PlatformUsername,
		APIGWDomain:      conf.APIGateway.Domain,
	})
	svc.cmdbHandler = cmdb.NewHandler(&cmdb.Config{
		Environment: conf.CMDB.Environment,
	}, svc.apigwCli)

	svc.topoStorage = topoStorage.NewStorage(&topoStorage.StorageConfig{
		MongoDB:            conf.MongoDB,
		Database:           TopoStorageDatabase,
		BusinessCollection: TopoStorageBusinessCollection,
		HostCollection:     TopoStorageHostCollection,
	})

	svc.manager = manager.NewManager(&manager.Config{
		Redis:   conf.Redis,
		MongoDB: conf.MongoDB,
	}, svc.cmdbHandler, svc.topoStorage)

	svc.Capability = &options.Capability{

		Manager:     svc.manager,
		TopoStorage: svc.topoStorage,
		CmdbHandler: svc.cmdbHandler,
		ApigwCli:    svc.apigwCli,
	}

	httpServer := router.NewRouter(
		svc.ctx,
		"http-server", conf.HTTPServer.BindIP, conf.HTTPServer.Port,
		svc.Capability,
		router.WithApiV3(),
		router.WithBasic(),
	)
	svc.routers = append(svc.routers, httpServer)

	return svc
}

// Start starts the backend service.
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

	if err := svc.topoStorage.Start(ctx); err != nil {
		blog.Errorf("failed to start topo storage, err: %v", err)

		return err
	}

	if err := svc.manager.Start(ctx); err != nil {
		blog.Errorf("failed to start manager, err: %v", err)

		return err
	}

	// start routers
	gp := gopool.NewPool()
	for _, router := range svc.routers {
		// router start will block until router stop, so we need to run it in a goroutine.
		fn := func() error {
			if err := router.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
		blog.Infof("start router,name: %s, ip: %s, port: %d",
			router.Name(), router.IP(), router.Port())
	}

	// wait until all routers stopped or application error.
	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to start routers, err: %v", err)
		return err
	}

	return nil
}
