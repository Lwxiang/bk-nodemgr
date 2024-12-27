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

	"git.woa.com/bk-gse/bk-nodeman/config"
	"git.woa.com/bk-gse/bk-nodeman/internal/manager"
	topoStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage/topo"
	"git.woa.com/bk-gse/bk-nodeman/pkg/apigw"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/cmdb"
)

// Service defines a server to provide all backend service.
type Service struct {
	conf *config.BackendService

	ctx        context.Context
	cancelFunc context.CancelFunc

	apigwCli    apigw.Client
	cmdbHandler cmdb.Handler
	topoStorage *topoStorage.DefaultStorage
	manager     *manager.Manager
}

// NewService creates a new backend service.
func NewService(conf *config.BackendService) *Service {
	apigwCli := apigw.NewClient(&apigw.Config{
		BKAppCode:        conf.APIGateway.AppCode,
		BKAppSecret:      conf.APIGateway.AppSecret,
		PlatformUsername: conf.APIGateway.PlatformUsername,
		APIGWDomain:      conf.APIGateway.Domain,
	})

	ch := cmdb.NewHandler(&cmdb.Config{
		Environment: "prod",
	}, apigwCli)

	ts := topoStorage.NewStorage(&topoStorage.StorageConfig{
		MongoDB:            conf.MongoDB,
		Database:           "nodeman",
		BusinessCollection: "business",
		HostCollection:     "host",
	})

	mgr := manager.NewManager(&manager.Config{
		Redis:   conf.Redis,
		MongoDB: conf.MongoDB,
	}, ch, ts)

	return &Service{
		conf:        conf,
		apigwCli:    apigwCli,
		cmdbHandler: ch,
		topoStorage: ts,
		manager:     mgr,
	}
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

	return nil
}
