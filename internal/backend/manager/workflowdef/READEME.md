## 文件定位

| 文件名                | 功能描述                | 备注               |
|--------------------|---------------------|------------------|
| action_def.go      | 存储所有 action_def 的名称 | 需要找 action 从此处开始 |
| oper_def.go        |                     |                  |
| {{action_name}}.go |                     |                  |

## action 模板

```go
/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflowdef ...
package workflowdef

import (
	"context"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/storage/topo"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/conv"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/tenant"
	"git.woa.com/bk-gse/bk-nodeman/pkg/thirdparty/cmdb"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
)

// NewActionActionName ...
func NewActionActionName() operengine.ActionDef {
	return &actionName{}
}

// ActionNameParam ...
type ActionNameParam struct {
}

// syncHostFromCMDB ...
type actionName struct {
}

// Name returns the name of the action.
func (s *actionName) Name() string {
	return ""
}

// Version returns the version of the action.
func (s *actionName) Version() string {
	return ""
}

// Description returns the description of the action.
func (s *actionName) Description() string {
	return ""
}

// Timeout returns the timeout of the action.
func (s *actionName) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (s *actionName) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (s *actionName) MaxRetryCount() uint {
	return 3
}

// DelayFn this func define when this action fails, how long to wait before retrying.
func (s *actionName) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do this func define what the action will do.
func (s *actionName) Do(ctx *operengine.ActionInstContext) error {
	param := new(ActionNameParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	/*
		do something
	*/

	return nil
}
```