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

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/manager/operationdef"
	"git.woa.com/bk-gse/bk-nodeman/internal/backend/manager/operationdef/actiondef"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
)

// Manager defines the manager interface.
type Manager interface {
	// Start starts the manager.
	Start(ctx context.Context) error

	// CheckHealth checks the health of manager.
	CheckHealth() error

	// GracefulShutdown ...
	GracefulShutdown() error

	// StartPipeline starts all reserved pipelines.
	StartPipeline(name operationdef.Name, timeout time.Duration) error
}

// NewManager creates a new manager.
func NewManager(conf Config, logger logger.Logger) (Manager, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	mgr := &manager{
		logger:         logger,
		isRunning:      false,
		operInstEngine: nil,
		conf:           conf,
	}

	var err error
	mgr.operInstEngine, err = operengine.NewOperInstEngine(
		1,
		operengine.WithRedis(
			mgr.conf.WorkflowConfig.Redis.Addr,
			mgr.conf.WorkflowConfig.Redis.Password,
			mgr.conf.WorkflowConfig.Redis.DB),
		mgr.conf.OperInstStorage,
		operengine.WithLogger(mgr.logger))
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

// manager provides to operate nodeman tasks.
type manager struct {
	logger logger.Logger

	// state
	isRunning bool

	operInstEngine operengine.OperInstEngine

	// config
	conf Config
}

// Start starts the manager.
func (mgr *manager) Start(ctx context.Context) error {
	if mgr.isRunning {
		return errors.New("manager already started")
	}

	if ctx == nil {
		return errors.New("context is nil")
	}

	if err := mgr.conf.Validate(); err != nil {
		return fmt.Errorf("config is invalid, err: %v", err)
	}

	if err := mgr.startOperEngineManager(ctx); err != nil {
		return err
	}

	mgr.isRunning = true

	mgr.logger.Info("successfully started manager")

	return nil
}

// CheckHealth checks the health of manager.
func (mgr *manager) CheckHealth() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if err := mgr.conf.TopoStorage.CheckHealthz(); err != nil {
		return fmt.Errorf("topo storage is unhealthy, err: %v", err)
	}

	if mgr.operInstEngine == nil {
		return errors.New("task engine manager is not initialized")
	}

	if err := mgr.operInstEngine.CheckHealth(); err != nil {
		return fmt.Errorf("operation instance engine manager is unhealthy, err: %v", err)
	}

	return nil
}

// GracefulShutdown ...
func (mgr *manager) GracefulShutdown() error {
	if !mgr.isRunning {
		return errors.New("manager is not running")
	}

	if err := mgr.operInstEngine.GracefulShutdown(); err != nil {
		return err
	}

	return nil
}

func (mgr *manager) startOperEngineManager(ctx context.Context) error {
	// TODO: implement me
	if err := mgr.registerActionDefs(); err != nil {
		return err
	}

	if err := mgr.operInstEngine.Start(ctx); err != nil {
		return err
	}

	return nil
}

// registerActionDefs init action defs
func (mgr *manager) registerActionDefs() error {
	return mgr.operInstEngine.RegisterActions(
		actiondef.NewActionSyncBusinessFromCMDB(mgr.conf.CmdbHandler, mgr.conf.TopoStorage),
	)
}

// StartPipeline start a pre-defined pipeline
func (mgr *manager) StartPipeline(name operationdef.Name, timeout time.Duration) error {
	task, err := operationdef.Factory()[name](mgr.operInstEngine).NewInstance(timeout)
	if err != nil {
		return err
	}

	if err := mgr.operInstEngine.DispatchOperationInst(task); err != nil {
		return err
	}

	return nil
}
