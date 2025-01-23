/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operationdef ...
package operationdef

import (
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
)

// Name is the pipeline name.
type Name string

const (
	SyncFromCmdb Name = "sync-from-cmdb"
)

// Factory ...
func Factory() map[Name]func(mgr operengine.OperInstEngine) *operengine.OperationDef {
	return map[Name]func(mgr operengine.OperInstEngine) *operengine.OperationDef{
		SyncFromCmdb: syncingFromCMDB,
	}
}
