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
	"errors"
	"fmt"
	"time"
)

// ActionInstData action instance.
type ActionInstData struct {
	TriggerID  string
	OperInstID string
	Name       string
	Index      int
	State      ActionInstState
	StartedAt  time.Time
	EndedAt    time.Time
	StoppedAt  time.Time
	Messages   []string
	Content    map[string]any
}

// Info ...
func (data *ActionInstData) Info() string {
	return fmt.Sprintf("operation-inst-id(%s), index(%d), action-name(%s)", data.OperInstID, data.Index, data.Name)
}

// Log log messages.
func (data *ActionInstData) Log(messages ...string) {
	data.Messages = append(data.Messages, messages...)
}

// OperInstData OperInst data.
type OperInstData struct {
	OperInstID        string
	OperationDefName  string
	ActionNames       []string
	ActionInstDataMap map[string]*ActionInstData
	ParentOperInstID  string
	Timeout           time.Duration

	InitContent map[string]map[string]any

	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
	StoppedAt time.Time
}

// OperInst is a operationDef instance.
type OperInst struct {
	operationDef *operationDef
	data         *OperInstData
}

// Validate  the operation.
func (o *OperInst) Validate() error {
	if o == nil {
		return errors.New("operation is nil")
	}

	if o.operationDef.name == "" {
		return errors.New("operationDef name is empty")
	}

	return nil
}

// LastActionInstState get the last action State.
func (o *OperInst) LastActionInstState(actionName string) ActionInstState {
	lastActionInstState := ActionInstStateSuccess

	for _, action := range o.data.ActionNames {
		if action != actionName {
			lastActionData, ok := o.data.ActionInstDataMap[action]
			if !ok {
				return ActionInstanceStateUnknown
			}

			lastActionInstState = lastActionData.State
			continue
		}
	}

	return lastActionInstState
}

// GetActionInstData get the action instance.
func (o *OperInst) GetActionInstData(actionName string) (*ActionInstData, error) {
	actionData, ok := o.data.ActionInstDataMap[actionName]
	if !ok {
		return nil, fmt.Errorf("action not found, name(%s)", actionName)
	}

	return actionData, nil
}
