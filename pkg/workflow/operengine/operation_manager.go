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
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/locker"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
)

// ErrOperationProcessing define an error if operation is processing.
func ErrOperationProcessing() error {
	return errors.New("operation is processing")
}

// OperationManager defines the operation manager.
type OperationManager interface {
	// Terminate an operation.
	Terminate(operationID string) error

	// Execute an operation.
	Execute(ctx context.Context, operationID string, param *OperationInstParam) error

	// Pause an operation.
	Pause(operationID string) error

	// Resume an operation.
	Resume(operationID string) error
}

// operationManager ...
type operationManager struct {
	operationInstEngine OperInstEngine
	storage             OperationStorage
	mutexFactory        locker.MutexFactory
	logger              logger.Logger
}

// AcquireTimeout is the maximum time to wait for the mutexFactory.
const AcquireTimeout = 5 * time.Second

// createOperationInst create an operation instance.
func (m *operationManager) createOperationInst(operation *Operation, param *OperationInstParam) (
	*OperationInst, error) {

	operationDef := NewOperationDef(operation.defSnapshot.OperationDefName)
	for _, actionName := range operation.defSnapshot.ActionNames {
		operationDef.Next(m.operationInstEngine.GetRegisteredAction(actionName))
	}

	operationInst, err := operationDef.NewInstance(param.Timeout)
	if err != nil {
		return nil, err
	}

	return operationInst, nil
}

// Terminate an operation.
func (m *operationManager) Terminate(operationID string) (err error) {
	mutex := m.mutexFactory.NewMutex(operationID)
	if err = mutex.TryLock(); err != nil {
		return err
	}

	defer func() {
		err = mutex.Unlock()
	}()

	operation, err := m.storage.GetOperation(operationID)
	if err != nil {
		return err
	}

	if err := m.operationInstEngine.Terminate(operation.getLatestOperationInstID()); err != nil {
		return err
	}

	operation.state = OperationStateTerminated
	if err = m.storage.UpdateOperation(operation); err != nil {
		return err
	}

	return nil
}

// Execute an operation.
func (m *operationManager) Execute(ctx context.Context, operationID string, param *OperationInstParam) (err error) {
	mutex := m.mutexFactory.NewMutex(operationID)
	if err = mutex.TryLock(); err != nil {
		return err
	}

	defer func() {
		err = mutex.Unlock()
	}()

	operation, err := m.storage.GetOperation(operationID)
	if err != nil {
		return err
	}

	if err = operation.CheckEnforceability(); err != nil {
		return err
	}

	operationInst, err := m.createOperationInst(operation, param)
	if err != nil {
		return err
	}

	operation.operationInstIDs = append(operation.operationInstIDs, operationInst.OperInstID)
	if err = m.storage.UpdateOperation(operation); err != nil {
		return err
	}

	if err = m.operationInstEngine.DispatchOperationInst(operationInst); err != nil {
		return err
	}

	return nil
}

// Pause an operation.
func (m *operationManager) Pause(operationID string) error {
	//TODO implement me
	panic("implement me")
}

// Resume an operation.
func (m *operationManager) Resume(operationID string) error {
	//TODO implement me
	panic("implement me")
}
