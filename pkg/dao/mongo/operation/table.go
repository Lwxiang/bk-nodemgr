/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operation ...
package operation

import (
	"strconv"

	"git.woa.com/bk-gse/bk-nodeman/pkg/dao/mongo/base"
)

// TableName operation table name.
const TableName = "operation"

// Operation represents an operation under a tenant.
// BizID should be the unique key.
type Operation struct {
	OperationID int64    `json:"operation_id" bson:"operation_id"`
	TriggerID   string   `json:"trigger_id" bson:"trigger_id"`
	OperInstIDs []string `json:"oper_inst_ids" bson:"oper_inst_ids"`
	DefSnapshot DefSnapshot
	State       string `json:"state" bson:"state"`
}

// DefSnapshot represents the snapshot of the operation definition.
type DefSnapshot struct {
	OperationDefName string   `json:"pipeline_name"`
	ActionNames      []string `json:"action_names"`
}

// UniqueKey unique key of the table.
func (biz *Operation) UniqueKey() string {
	return strconv.FormatInt(biz.OperationID, 10)
}

// TableOperation represents the complete db structures of an operation.
type TableOperation base.TableBroker[*Operation]
