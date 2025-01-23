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
	"fmt"
	"time"
)

const (
	// OperationMaxInstNum define a max instance number of an operation.
	OperationMaxInstNum = 100
)

// Operation ...
type Operation struct {
	TriggerID        string
	OperationID      string
	defSnapshot      OperationDefSnapshot
	operationInstIDs []string
	state            OperationState
}

// Validate the operation.
func (operation *Operation) Validate() error {
	// TODO: Validate the operation.
	return nil
}

// CheckEnforceability check if an operation can be executed.
func (operation *Operation) CheckEnforceability() error {
	// check the operation instance length.
	if len(operation.operationInstIDs) > OperationMaxInstNum {
		return fmt.Errorf("operation can not be executed, operation-inst-length(%d)", len(operation.operationInstIDs))
	}

	// check the state of operation.
	switch operation.state {
	// in this case, operation can be executed.
	case OperationStateInit, OperationStateRunning, OperationStateFailed:
	// in default, operation can not be executed.
	default:
		return fmt.Errorf("operation can not be executed, state(%s)", operation.state)
	}

	return nil
}

// getLatestOperationInstID ...
func (operation *Operation) getLatestOperationInstID() string {
	if len(operation.operationInstIDs) == 0 {
		return ""
	}

	return operation.operationInstIDs[len(operation.operationInstIDs)-1]
}

// OperationDefSnapshot defines the snapshot of operationDef.
type OperationDefSnapshot struct {
	OperationDefName string   `json:"pipeline_name"`
	ActionNames      []string `json:"action_names"`
}

// OperationInstParam ...
type OperationInstParam struct {
	// Timeout define the timeout of operation instance.
	Timeout  time.Duration
	Metadata map[string]any
}
