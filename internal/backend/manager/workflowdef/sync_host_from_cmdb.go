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

// NewActionSyncHostFromCMDB ...
func NewActionSyncHostFromCMDB(cmdbHandler cmdb.Handler, topoStorage topo.Storage) operengine.ActionDef {
	return &syncHostFromCMDB{
		cmdbHandler: cmdbHandler,
		topoStorage: topoStorage,
	}
}

// syncHostFromCMDBParam ...
type syncHostFromCMDBParam struct {
	BizID    int64  `json:"biz_id"`
	TenantID string `json:"tenant_id"`
}

// syncHostFromCMDB ...
type syncHostFromCMDB struct {
	cmdbHandler cmdb.Handler
	topoStorage topo.Storage
}

// Name ...
func (s *syncHostFromCMDB) Name() string {
	return SyncHostFromCMDB
}

// Version ...
func (s *syncHostFromCMDB) Version() string {
	return "v1"
}

// Description ...
func (s *syncHostFromCMDB) Description() string {
	return "Get the host information of the designated business from CMDB, and update to the database."
}

// Timeout ...
func (s *syncHostFromCMDB) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags ...
func (s *syncHostFromCMDB) Tags() []operengine.ActionTag {
	return []operengine.ActionTag{}
}

// MaxRetryCount ...
func (s *syncHostFromCMDB) MaxRetryCount() uint {
	return 3
}

// DelayFn ...
func (s *syncHostFromCMDB) DelayFn() func() {
	return func() {
		time.Sleep(1 * time.Second)
	}
}

// Do ...
func (s *syncHostFromCMDB) Do(ctx *operengine.ActionInstContext) error {
	param := new(syncHostFromCMDBParam)
	err := conv.MapToStruct(ctx.Data.Content, param)
	if err != nil {
		return err
	}

	executor := runtime.NewPageExecutor[*types.Host](500, 1*time.Hour)
	fn := func(ctx context.Context, p types.Page) ([]*types.Host, error) {
		ctx, err = tenant.SetID(ctx, param.TenantID)
		if err != nil {
			return nil, err
		}

		hosts, err := s.cmdbHandler.ListBizHosts(ctx, param.BizID, p)
		if err != nil {
			return nil, err
		}

		return hosts, nil
	}

	result, err := executor.Execute(ctx.Ctx, types.UnlimitedPage(), fn)
	if err != nil {
		return err
	}

	if err = s.topoStorage.UpsertHosts(ctx.Ctx, result.Items...); err != nil {
		return err
	}

	return nil
}
