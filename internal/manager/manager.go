/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package manager provides handlers to manage all the nodeman task operations
package manager

import (
	"context"
	"errors"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/config"
	"git.woa.com/bk-gse/bk-nodeman/internal/manager/actions"
	topoStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage/topo"
	workflowStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage/workflow"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/cmdb"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow"
)

// NewManager creates a new manager.
func NewManager(config *Config, cmdbHandler cmdb.Handler, topoStorage topoStorage.Storage) *Manager {
	return &Manager{
		config:      config,
		cmdbHandler: cmdbHandler,
		topoStorage: topoStorage,
	}
}

// Config defines the configuration of manager.
type Config struct {
	Redis   config.Redis
	MongoDB config.MongoDB
}

// Manager provides to operate nodeman tasks.
type Manager struct {
	// config
	config *Config

	// state
	isRunning bool

	workflowMgr *workflow.Manager
	cmdbHandler cmdb.Handler
	topoStorage topoStorage.Storage
}

// Start starts the manager.
func (mgr *Manager) Start(ctx context.Context) error {
	if mgr.isRunning {
		return errors.New("manager already started")
	}

	if err := mgr.initializeWorkflowManager(ctx); err != nil {
		return err
	}

	mgr.isRunning = true

	go mgr.startSyncingFromCMDB()
	blog.Info("successfully started manager")

	return nil
}

func (mgr *Manager) initializeWorkflowManager(ctx context.Context) error {
	workflowStg := workflowStorage.NewStorage(&workflowStorage.StorageConfig{
		MongoDB:            mgr.config.MongoDB,
		Database:           "workflow",
		TaskCollection:     "task",
		StoppingCollection: "stopping_task",
	})

	if err := workflowStg.Start(ctx); err != nil {
		return err
	}

	mgr.workflowMgr = workflow.NewManager(
		&workflow.ManagerConfig{
			Redis:     mgr.config.Redis,
			MongoDB:   mgr.config.MongoDB,
			WorkerNum: 1,
		}, workflowStg)

	if err := mgr.initializeActionDefs(); err != nil {
		return err
	}

	if err := mgr.workflowMgr.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (mgr *Manager) initializeActionDefs() error {
	return mgr.workflowMgr.RegisterActions(
		actions.NewActionSyncBusinessFromCMDB(mgr.cmdbHandler, mgr.topoStorage),
	)
}

func (mgr *Manager) startSyncingFromCMDB() {
	task, err := workflow.NewPipeline("sync-from-cmdb").
		Next(mgr.workflowMgr.GetRegisteredAction(actions.ActionNameSyncBusinessFromCMDB)).NewPeriodTask(10*time.Minute, "*/1 * * * *")
	if err != nil {
		blog.Warnf("failed to create syncing from cmdb task: %v", err)

		return
	}

	if err := mgr.workflowMgr.DispatchPeriodTask(task); err != nil {
		blog.Warnf("failed to dispatch syncing from cmdb task: %v", err)

		return
	}

	blog.Info("successfully started period task keep syncing from cmdb")
}
