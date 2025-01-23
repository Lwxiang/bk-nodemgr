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

	"github.com/google/uuid"
)

// OperationInstTimeoutDefault default OperationInst timeout.
const OperationInstTimeoutDefault = 10 * time.Minute

// NewOperationDef creates a new OperationDef.
func NewOperationDef(name string) *OperationDef {
	return &OperationDef{
		name:       name,
		actionDefs: make([]ActionDef, 0),
	}
}

// OperationDef represents a operationDef definition.
type OperationDef struct {
	name       string
	actionDefs []ActionDef
}

// Name returns the name of the OperationDef.
func (def *OperationDef) Name() string {
	return def.name
}

// Next appends a new action to the OperationDef.
func (def *OperationDef) Next(actionDef ActionDef) *OperationDef {
	def.actionDefs = append(def.actionDefs, actionDef)

	return def
}

// OperationInstIDPrefix ...
const OperationInstIDPrefix = "operation-inst"

// NewInstance creates a new OperationInst.
func (def *OperationDef) NewInstance(timeout time.Duration) (*OperationInst, error) {
	if err := def.Validate(); err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = OperationInstTimeoutDefault
	}

	inst := &OperationInst{
		data: &OperationInstData{
			OperInstID:        fmt.Sprintf("%s-%s", OperationInstIDPrefix, uuid.NewString()),
			OperationDefName:  def.name,
			ActionNames:       make([]string, len(def.actionDefs)),
			ActionInstDataMap: make(map[string]*ActionInstData),
			Timeout:           timeout,
			CreatedAt:         time.Now().Local(),
		},
		operationDef: def,
	}

	for index, action := range def.actionDefs {
		inst.data.ActionNames[index] = action.Name()
		inst.data.ActionInstDataMap[action.Name()] = &ActionInstData{
			OperInstID: inst.data.OperInstID,
			Name:       action.Name(),
			Index:      index,
			State:      ActionInstanceStatePending,
			Messages:   make([]string, 0),
			Content:    "",
		}
	}

	return inst, nil
}

// Validate validates the OperationDef.
func (def *OperationDef) Validate() error {
	if len(def.actionDefs) == 0 {
		return errors.New("empty operationDef definition")
	}

	for i, action := range def.actionDefs {
		if action == nil {
			return fmt.Errorf("action is nil, index(%d)", i)
		}
	}

	return nil
}
