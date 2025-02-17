/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package actions is the actions for workflow manager.
package actions

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/cmdb"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// NewActionSyncBusinessFromCMDB creates a new ActionSyncBusinessFromCMDB.
func NewActionSyncBusinessFromCMDB(cmdbHandler cmdb.Handler, topoStorage topo.Storage) *ActionSyncBusinessFromCMDB {
	return &ActionSyncBusinessFromCMDB{
		cmdbHandler: cmdbHandler,
		topoStorage: topoStorage,
	}
}

const (
	// ActionNameSyncBusinessFromCMDB defines the name of this action.
	ActionNameSyncBusinessFromCMDB = "sync_business_from_cmdb"
)

// ActionSyncBusinessFromCMDB sync business info from cmdb.
type ActionSyncBusinessFromCMDB struct {
	cmdbHandler cmdb.Handler
	topoStorage topo.Storage
}

// Name returns the name of the action.
func (a *ActionSyncBusinessFromCMDB) Name() string {
	return ActionNameSyncBusinessFromCMDB
}

// Version returns the version of the action.
func (a *ActionSyncBusinessFromCMDB) Version() string {
	return "v1"
}

// Description returns the description of the action.
func (a *ActionSyncBusinessFromCMDB) Description() string {
	return "sync business info from cmdb and update to storage"
}

// Timeout returns the timeout of this action.
func (a *ActionSyncBusinessFromCMDB) Timeout() time.Duration {
	return 1 * time.Minute
}

// MaxRetryCount returns the max retry count of this action.
func (a *ActionSyncBusinessFromCMDB) MaxRetryCount() uint {
	return 0
}

// Do executes the action.
func (a *ActionSyncBusinessFromCMDB) Do(ctx *operengine.ActionInstContext) error {
	blog.Infof("start syncing business info from cmdb. info: %s", ctx.Data.Info())
	ctx.Data.Log("start syncing business info from cmdb")

	gp := gopool.NewPool()
	gp.SetLimit(10)

	page := types.Page{
		Offset: 0,
		Limit:  500,
	}
	for {
		businesses, err := a.cmdbHandler.SearchBusiness(context.Background(), page)
		if err != nil {
			blog.Errorf("failed to get business info from cmdb, err: %v", err)
			ctx.Data.Log("failed to get business info from cmdb. err: " + err.Error())

			return err
		}

		if len(businesses) == 0 {
			break
		}

		for idx := range businesses {
			business := businesses[idx]
			fn := func() error {
				if err := a.topoStorage.UpsertBusiness(ctx.Ctx, &business); err != nil {
					return err
				}

				return nil
			}

			gp.Go(fn)
		}

		page.Offset += page.Limit
	}

	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to sync business info from cmdb. info: %s, err: %v", ctx.Data.Info(), err)
		return err
	}

	blog.Infof("succeed to sync business info from cmdb. info: %s", ctx.Data.Info())

	return nil
}
