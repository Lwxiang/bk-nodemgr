/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operdef ...
package workflowdef

import (
	"fmt"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/storage/topo"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/conv"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/tenant"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
)

// NewActionGenAllBizHostSyncOper this action will create host sync operation for all business.
func NewActionGenAllBizHostSyncOper(topoStorage topo.Storage, operMgr operengine.OperationMgr) operengine.ActionDef {
	return &genAllBizHostSyncOper{
		topoStorage: topoStorage,
		operMgr:     operMgr,
	}
}

// genAllBizHostSyncOper ...
type genAllBizHostSyncOper struct {
	topoStorage topo.Storage
	operMgr     operengine.OperationMgr
}

// Name ...
func (c *genAllBizHostSyncOper) Name() string {
	return GenAllBizHostSyncOper
}

// Version ...
func (c *genAllBizHostSyncOper) Version() string {
	return "v1.0.0"
}

// Description ...
func (c *genAllBizHostSyncOper) Description() string {
	return "reads all business information from the database," +
		"and creates host synchronization tasks on a business-by-business basis."
}

// Timeout ...
func (c *genAllBizHostSyncOper) Timeout() time.Duration {
	return time.Second * 10
}

// Tags ...
func (c *genAllBizHostSyncOper) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount this action creates a large number of synchronization tasks,
// therefore does not allow the system to automatically retry.
func (c *genAllBizHostSyncOper) MaxRetryCount() uint {
	return 0
}

// DelayFn ...
func (c *genAllBizHostSyncOper) DelayFn() func() {
	return func() {}
}

// Do ...
func (c *genAllBizHostSyncOper) Do(ctx *operengine.ActionInstContext) error {
	ctx.Data.Log("successfully start create all biz host sync operation")
	// TODO: 支持多租户版本
	tenantID := "0"
	tenantCtx, err := tenant.SetID(ctx.Ctx, tenantID)
	if err != nil {
		return err
	}

	bizs, err := c.topoStorage.ListBusinesses(tenantCtx)
	if err != nil {
		return err
	}

	ctx.Data.Log(fmt.Sprintf("found %d business", len(bizs)))

	for idx, _ := range bizs {
		biz := bizs[idx]

		ctx.Data.Log(fmt.Sprintf("start create sync host operation for business, biz-name(%s), biz-id(%d)",
			biz.BizName, biz.BizID))

		operation := NewOperSyncHostFromCMDB(ctx.Data.TriggerID)
		err := c.operMgr.ExecuteOperation(operation, &operengine.OperInstParam{
			Timeout: time.Second * 10,
			InitContent: map[string]map[string]any{
				SyncHostFromCMDB: conv.StructToMapIgnoreError(syncHostFromCMDBParam{
					BizID:    biz.BizID,
					TenantID: tenantID,
				}),
			},
		})
		if err != nil {
			ctx.Data.Log(fmt.Sprintf("failed to create sync host operation for business, biz-name(%s), biz-id(%d)",
				biz.BizName, biz.BizID))
			return err
		}
	}

	ctx.Data.Log(fmt.Sprintf("successfully create all biz host sync operation"))

	return nil
}
