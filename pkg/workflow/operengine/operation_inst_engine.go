/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operengine ...
package operengine

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"github.com/RichardKnop/machinery/v2"
	backendIface "github.com/RichardKnop/machinery/v2/backends/iface"
	redisBackend "github.com/RichardKnop/machinery/v2/backends/redis"
	brokerIface "github.com/RichardKnop/machinery/v2/brokers/iface"
	redisBroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/tasks"
)

const (
	// DefaultQueueName defines the default queue name.
	DefaultQueueName = "operation_inst_engine_queue"

	// ResultsExpireInDefault defines the default results expire in.
	ResultsExpireInDefault = 3600
)

// OperInstEngine defines the operation engine.
type OperInstEngine interface {
	// Start ...
	Start(ctx context.Context) error

	// CheckHealth ...
	CheckHealth() error

	// GracefulShutdown ...
	GracefulShutdown() error

	// DispatchOperationInst ...
	DispatchOperationInst(operation *OperationInst) error

	// GetRegisteredAction ...
	GetRegisteredAction(name string) ActionDef

	// RegisterActions ...
	RegisterActions(actionDefs ...ActionDef) error

	// Terminate an operation inst.
	Terminate(operationInstID string) error
}

// engine provides operation inst engine consumer and producer.
type engine struct {
	// machinery config
	mConfig *machineryConfig.Config

	// broker
	broker brokerIface.Broker

	// backend
	backend backendIface.Backend

	// WorkerNum defines the number of workers.
	WorkerNum int

	// state
	isRunning   bool
	isComsuming bool

	// context
	ctx    context.Context
	cancel context.CancelFunc

	server *machinery.Server
	worker *machinery.Worker

	storage OperationInstStorage

	logger logger.Logger

	registeredActionDefs map[string]ActionDef

	launchWorkerErr chan error
}

// OptionsFunc is a function that configures an engine.
type OptionsFunc func(m *engine)

// WithLogger sets the logger for the engine.
func WithLogger(logger logger.Logger) OptionsFunc {
	return func(m *engine) {
		m.logger = logger
	}
}

// ServerOptionFn will be used to set the server's backend and broker.
type ServerOptionFn func(m *engine)

const (
	redisMaxIdle                = 10
	redisTimeout                = 240
	redisReadTimeout            = 15
	redisWriteTimeout           = 15
	redisConnectTimeout         = 15
	redisNormalTasksPollPeriod  = 1000
	redisDelayedTasksPollPeriod = 500
)

// WithRedis sets the redis broker for the engine.
func WithRedis(address, password string, db int) ServerOptionFn {
	return func(e *engine) {
		e.mConfig.Redis = &machineryConfig.RedisConfig{
			MaxIdle:                redisMaxIdle,
			IdleTimeout:            redisTimeout,
			ReadTimeout:            redisReadTimeout,
			WriteTimeout:           redisWriteTimeout,
			ConnectTimeout:         redisConnectTimeout,
			NormalTasksPollPeriod:  redisNormalTasksPollPeriod,
			DelayedTasksPollPeriod: redisDelayedTasksPollPeriod,
		}

		e.broker = redisBroker.New(e.mConfig, address, password, "", db)
		e.backend = redisBackend.New(e.mConfig, address, password, "", db)
	}
}

// NewOperInstEngine creates a new OperationInst engine.
func NewOperInstEngine(workerNum int, envFunc ServerOptionFn, storage OperationInstStorage, opts ...OptionsFunc) (
	OperInstEngine, error) {

	e := &engine{
		mConfig: &machineryConfig.Config{
			DefaultQueue:    DefaultQueueName,
			ResultsExpireIn: ResultsExpireInDefault,
			NoUnixSignals:   true,
		},
		isRunning:            false,
		registeredActionDefs: make(map[string]ActionDef),
		storage:              storage,
		logger:               logger.LoggerDefault{},
		WorkerNum:            workerNum,
		launchWorkerErr:      make(chan error, 1),
	}

	envFunc(e)

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

// Start starts the OperationInst engine.
func (e *engine) Start(ctx context.Context) error {
	if e.isRunning {
		return errors.New("engine already started")
	}

	e.ctx, e.cancel = context.WithCancel(ctx)

	if err := e.initialize(); err != nil {
		return err
	}

	if e.WorkerNum > 0 {
		if err := e.launchWorker(); err != nil {
			return err
		}
	}

	return nil
}

// GracefulShutdown shuts down the engine gracefully.
func (e *engine) GracefulShutdown() error {
	if !e.isRunning {
		return errors.New("engine is not running")
	}

	if e.isComsuming {
		go e.broker.StopConsuming()
	}

	e.isRunning = false
	e.isComsuming = false

	defer e.cancel()

	err := <-e.launchWorkerErr
	if !errors.Is(err, machinery.ErrWorkerQuitGracefully) {
		return err
	}

	return nil
}

// CheckHealth checks the health of the engine.
func (e *engine) CheckHealth() error {
	if !e.isRunning {
		return errors.New("engine is not running")
	}

	if e.server == nil {
		return errors.New("machinery server is not initialized")
	}

	if e.worker == nil {
		return errors.New("worker is not initialized")
	}

	return nil
}

// RegisterAction registers a OperationInst engine action.
func (e *engine) RegisterAction(actionDef ActionDef) error {
	if e.isRunning {
		return errors.New("engine already started, can not register action")
	}

	if _, ok := e.registeredActionDefs[actionDef.Name()]; ok {
		return errors.New("action already registered")
	}

	e.registeredActionDefs[actionDef.Name()] = actionDef

	return nil
}

// RegisterActions registers multiple OperationInst engine actions.
func (e *engine) RegisterActions(actionDefs ...ActionDef) error {
	for _, actionDef := range actionDefs {
		if err := e.RegisterAction(actionDef); err != nil {
			return err
		}
	}

	return nil
}

// GetRegisteredAction returns a registered OperationInst engine action.
func (e *engine) GetRegisteredAction(name string) ActionDef {
	return e.registeredActionDefs[name]
}

// DispatchOperationInst dispatches an operation inst to the engine.
func (e *engine) DispatchOperationInst(inst *OperationInst) error {
	if !e.isRunning {
		return errors.New("engine is not running")
	}

	if err := inst.Validate(); err != nil {
		return err
	}

	if err := e.storeOperationInst(inst); err != nil {
		return err
	}

	return e.dispatchOperationInst(inst)
}

// StopOperationInst stops an operation inst.
func (e *engine) StopOperationInst(operationInstID string) error {
	return e.storage.MarkOperationInstStopping(nil, operationInstID)
}

// initialize initializes the OperationInst engine.
func (e *engine) initialize() error {
	if e.WorkerNum <= 0 {
		return errors.New("worker num should be greater than 0")
	}

	if e.mConfig == nil {
		return errors.New("engine config is nil")
	}

	if e.broker == nil {
		return errors.New("broker is nil")
	}

	if e.backend == nil {
		return errors.New("backend is nil")
	}

	if e.storage == nil {
		return errors.New("storage is nil")
	}

	// We don't need to use the periodic task of machinery, so there is no need to access the lock.
	e.server = machinery.NewServer(e.mConfig, e.broker, e.backend, nil)

	e.isRunning = true

	return nil
}

const consumerTag = ""

func (e *engine) launchWorker() error {
	if !e.isRunning {
		return errors.New("engine is not running")
	}

	for _, actionDef := range e.registeredActionDefs {
		if err := e.server.RegisterTask(actionDef.Name(), e.do); err != nil {
			return err
		}
	}

	e.worker = e.server.NewWorker(consumerTag, e.WorkerNum)

	e.worker.SetErrorHandler(func(err error) {
		e.logger.Errorf("worker error, err: %v", err)
		e.logger.Debugf("stack: %s", string(debug.Stack()))
	})
	e.worker.SetPreTaskHandler(func(signature *tasks.Signature) {})
	e.worker.SetPostTaskHandler(func(signature *tasks.Signature) {})

	e.isComsuming = true
	go func() {
		e.launchWorkerErr <- e.worker.Launch()
		e.isComsuming = false
	}()

	return nil
}

// dispatchOperationInst dispatch OperationInst to machinery chain.
func (e *engine) dispatchOperationInst(inst *OperationInst) error {
	var signatures []*tasks.Signature

	for _, actionDef := range inst.operationDef.actionDefs {
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("%s-%s", inst.data.OperInstID, actionDef.Name()),
			Name: actionDef.Name(),
			Args: []tasks.Arg{
				{
					Name:  "action",
					Type:  "string",
					Value: actionDef.Name(),
				},
				{
					Name:  "operation-instance-id",
					Type:  "string",
					Value: inst.data.OperInstID,
				},
			},
		}

		signatures = append(signatures, signature)
	}

	if len(signatures) == 0 {
		return fmt.Errorf("no action to do")
	}

	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		return err
	}

	_, err = e.server.SendChainWithContext(e.ctx, chain)
	if err != nil {
		return fmt.Errorf("send chain to machinery failed, err: %v", err)
	}

	return nil
}

// EngineMaxRetryLimit is the max retry limit of an action.
const EngineMaxRetryLimit = uint(10)

// do will dispatch the action of OperationInst to machinery.
func (e *engine) do(ctx context.Context, actionName string, operationInstID string) error {
	actionDef, ok := e.registeredActionDefs[actionName]
	if !ok {
		return fmt.Errorf("action not registered, name(%s)", actionName)
	}

	inst, err := e.getOperationInst(operationInstID)
	if err != nil {
		return err
	}

	data, err := inst.GetActionInstData(actionName)
	if err != nil {
		return fmt.Errorf("failed to get action instance data from operation inst. "+
			"operation-inst-id(%s), action-name(%s)", operationInstID, actionName)
	}

	skip, err := evaluateActionInstanceState(data)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	if err := e.flashActionInstData(data, inst); err != nil {
		return err
	}

	// action context is used to control the timeout of the action.
	actionCtx, actionCancel := context.WithTimeout(ctx, actionDef.Timeout())
	defer actionCancel()

	// operation inst context is used to control the timeout of the operation inst.
	startedAt := inst.data.StartedAt
	operationInstCtx, cancel := context.WithDeadline(ctx, startedAt.Add(inst.data.Timeout))
	defer cancel()

	// watch storage for stopping event.
	terminatingC := e.storage.WatchOperInstStopping(actionCtx, inst.data.OperInstID)

	doResult := make(chan error, 1)
	actionInstCtx := &ActionInstContext{
		Ctx:  actionCtx,
		Data: data,
	}

	go e.callActionDef(doResult, actionInstCtx, actionDef)

	select {
	case err = <-doResult:
		data.EndedAt = time.Now()

		if err == nil {
			data.State = ActionInstanceStateSuccess
		} else {
			data.State = ActionInstanceStateFailed
		}

		if storeErr := inst.store(); storeErr != nil {
			return fmt.Errorf("failed to store operation inst param. operation-inst-id(%s), action-name(%s), err: %v",
				operationInstID, actionName, storeErr)
		}

		return err

	case <-actionCtx.Done():
		return fmt.Errorf("action timeout, operation-inst-id(%s), action-name(%s)", operationInstID, actionName)
	case <-operationInstCtx.Done():
		return fmt.Errorf("operation inst timeout, operation-inst-id(%s), action-name(%s)",
			operationInstID, actionName)
	case <-terminatingC:
		return fmt.Errorf("operation inst has been terminated, operation-inst-id(%s), action-name(%s)",
			operationInstID, actionName)
	case <-ctx.Done():
		return fmt.Errorf("operation engine context done, operation-inst-id(%s), action-name(%s)",
			operationInstID, actionName)
	}
}

func (e *engine) flashActionInstData(data *ActionInstData, inst *OperationInst) error {
	nowTime := time.Now()
	if data.Index == 0 {
		inst.data.StartedAt = nowTime
	}
	data.StartedAt = nowTime
	data.State = ActionInstanceStateRunning

	if err := inst.store(); err != nil {
		return fmt.Errorf("failed to store operation inst param. operation-inst-id(%s), action-name(%s), err: %v",
			inst.OperInstID, data.Name, err)
	}

	return nil
}

// evaluateActionInstanceState evaluate action state is ready to run.
func evaluateActionInstanceState(data *ActionInstData) (bool, error) {
	if data == nil {
		return false, fmt.Errorf("action instance data is nil")
	}

	switch data.State {
	case ActionInstanceStateSuccess, ActionInstanceStateSkipped:
		return true, nil
	case ActionInstanceStateFailed, ActionInstanceStateTimeout,
		ActionInstanceStateTerminated:
		return false, fmt.Errorf("operation instance has completed. operation-inst-id(%s), action-name(%s), state(%s)",
			data.OperInstID, data.Name, data.State)
	case ActionInstanceStateRunning:
		return false, fmt.Errorf("action is running. operation-inst-id(%s), action-name(%s), state(%s)",
			data.OperInstID, data.Name, data.State)
	case ActionInstanceStatePending:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected action state. operation-inst-id(%s), action-name(%s), state(%s)",
			data.OperInstID, data.Name, data.State)
	}
}

// callActionDef call action def.
func (e *engine) callActionDef(doResult chan error, actionInstCtx *ActionInstContext, actionDef ActionDef) {
	defer func() {
		if r := recover(); r != nil {
			doResult <- fmt.Errorf("action panic,info(%v), revoer(%v), stack(%s)",
				actionInstCtx.Data.Info(), r, debug.Stack())
		}
	}()

	e.logger.Infof("action start, info(%s), description(%s)", actionInstCtx.Data.Info(), actionDef.Description())

	maxRetryNum := actionDef.MaxRetryCount()
	var doErr error
	for retryNum := uint(0); retryNum <= maxRetryNum && retryNum < EngineMaxRetryLimit; retryNum++ {
		doErr = actionDef.Do(actionInstCtx)
		if doErr != nil {
			actionDef.DelayFn()
			continue
		}

		break
	}

	e.logger.Infof("action done, info(%s), description(%s), err: %v",
		actionInstCtx.Data.Info(), actionDef.Description(), doErr)

	doResult <- doErr
}

// storeOperationInst store OperationInst param to storage.
func (e *engine) storeOperationInst(t *OperationInst) error {
	if err := e.storage.CreateOperationInstData(context.Background(), t.data); err != nil {
		return err
	}

	e.logger.Infof("successfully store operation inst param, operation-inst-id(%s)", t.data.OperInstID)

	return nil
}

// getOperationInst get OperationInst param from storage.
func (e *engine) getOperationInst(operationInstID string) (*OperationInst, error) {
	data, err := e.storage.GetOperInstData(context.Background(), operationInstID)
	if err != nil {
		return nil, err
	}

	operationDef := &OperationDef{name: data.OperationDefName}
	for _, actionName := range data.ActionNames {
		actionDef, ok := e.registeredActionDefs[actionName]
		if !ok {
			return nil, fmt.Errorf("action not registered, action-name(%s)", actionName)
		}

		operationDef = operationDef.Next(actionDef)
	}

	return &OperationInst{
		data:         data,
		operationDef: operationDef,
		storeFn:      e.storage.UpdateOperationInstData,
	}, nil
}

// Terminate an operation.
func (e *engine) Terminate(operationInstID string) error {
	// TODO: implement me
	panic("implement me")
}
