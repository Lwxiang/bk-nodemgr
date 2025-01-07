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
	"fmt"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/manager/actions"
	topoStorage "git.woa.com/bk-gse/bk-nodeman/internal/backend/storage/topo"
	workflowStorage "git.woa.com/bk-gse/bk-nodeman/internal/backend/storage/workflow"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"git.woa.com/bk-gse/bk-nodeman/pkg/thirdparty/cmdb"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow"
)

// IManager defines the manager interface.
type IManager interface {
	// CheckHealth checks the health of manager.
	CheckHealth() error

	// StartPipeline starts all reserved pipelines.
	StartPipeline(name PipelineName, timeout time.Duration) error
}

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

	blog.Info("successfully started manager")

	return nil
}

// CheckHealth checks the health of manager.
func (mgr *Manager) CheckHealth() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if mgr.cmdbHandler == nil {
		return errors.New("cmdb handler is not initialized")
	}

	if mgr.topoStorage == nil {
		return errors.New("topo storage is not initialized")
	}

	if err := mgr.topoStorage.CheckHealthz(); err != nil {
		return fmt.Errorf("topo storage is unhealthy: %v", err)
	}

	if mgr.workflowMgr == nil {
		return errors.New("workflow manager is not initialized")
	}

	if err := mgr.workflowMgr.CheckHealth(); err != nil {
		return fmt.Errorf("workflow manager is unhealthy: %v", err)
	}

	return nil
}

func (mgr *Manager) initializeWorkflowManager(ctx context.Context) error {
	workflowStg := workflowStorage.NewStorage(&workflowStorage.StorageConfig{
		MongoDB:            mgr.config.MongoDB,
		Database:           "nodeman",
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

// initializeActionDefs init action defs
func (mgr *Manager) initializeActionDefs() error {
	return mgr.workflowMgr.RegisterActions(
		actions.NewActionSyncBusinessFromCMDB(mgr.cmdbHandler, mgr.topoStorage),
	)
}

// startSyncingFromCMDB start a period task to keep syncing from cmdb
// TODO: implement
func (mgr *Manager) startSyncingFromCMDB() {
	task, err := pipelineFactory[PipelineSyncFromCmdb](mgr.workflowMgr).
		NewPeriodTask(10*time.Minute, "*/1 * * * *")
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

// StartPipeline start a pre-defined pipeline
func (mgr *Manager) StartPipeline(name PipelineName, timeout time.Duration) error {
	task, err := pipelineFactory[name](mgr.workflowMgr).NewTask(timeout)
	if err != nil {
		return err
	}

	if err := mgr.workflowMgr.DispatchTask(task); err != nil {
		return err
	}

	return nil
}
