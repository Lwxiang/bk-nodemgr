/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package business provides business dao operations.
package business

import (
	"context"
	"fmt"
	"sync"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/tenant"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler business handler interface.
type Handler interface {
	// ListAll list all business.
	ListAll(ctx context.Context) ([]*types.Business, error)

	// UpsertMany updates or inserts business.
	UpsertMany(ctx context.Context, bizs ...*types.Business) error
}

type handler struct {
	client *mongo.Database
	logger logger.Logger
	// daoMap stores dao's containing tenant information.
	// Do not edit the daoMap except with the tenantDao func.
	daoMap sync.Map
}

func (h *handler) tenantDao(tenantID string) *dao {
	if d, ok := h.daoMap.Load(tenantID); ok {
		return d.(*dao)
	}

	d, _ := h.daoMap.LoadOrStore(tenantID, newDao(tenantID, h.client, h.logger))

	// note: we can be sure that only the tenantDao func edit the daoMap,
	// so we can just use the type assertion here.
	return d.(*dao)
}

// New create a new business handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		client: client,
		logger: logger,
		daoMap: sync.Map{},
	}
}

// ListAll list all business.
func (h *handler) ListAll(ctx context.Context) ([]*types.Business, error) {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return nil, err
	}

	bizs, err := h.tenantDao(tenantID).listAll(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*types.Business, len(bizs))
	for idx, biz := range bizs {
		data[idx] = &types.Business{
			TenantID: biz.TenantID,
			BizID:    biz.BizID,
			BizName:  biz.BizName,
		}
	}

	return data, nil
}

// UpsertMany updates or inserts business.
func (h *handler) UpsertMany(ctx context.Context, bizs ...*types.Business) error {
	tenantID, err := tenant.GetID(ctx)
	if err != nil {
		return err
	}

	data := make([]*Business, len(bizs))
	for idx, biz := range bizs {
		data[idx] = &Business{
			TenantID: biz.TenantID,
			BizID:    biz.BizID,
			BizName:  biz.BizName,
		}

		if data[idx].TenantID != tenantID {
			return fmt.Errorf("tenantID not match, ctx-tenantID(%s), biz-tenantID(%s)", tenantID, biz.TenantID)
		}
	}

	if err := h.tenantDao(tenantID).upsertMany(ctx, data); err != nil {
		return err
	}

	return nil
}
