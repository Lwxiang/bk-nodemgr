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

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler business handler interface.
type Handler interface {
	Upsert(ctx context.Context, biz *types.Business) error
	ListAll(ctx context.Context) ([]*types.Business, error)
}

type handler struct {
	dao *dao
}

// New create a new business handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		dao: newDao(client, logger),
	}
}

// Upsert updates or inserts a business.
func (h *handler) Upsert(ctx context.Context, biz *types.Business) error {
	data := &Business{
		TenantID: biz.TenantID,
		BizID:    biz.BizID,
		BizName:  biz.BizName,
	}
	if err := h.dao.upsert(ctx, data); err != nil {
		return err
	}

	return nil
}

// ListAll list all business.
func (h *handler) ListAll(ctx context.Context) ([]*types.Business, error) {
	bizs, err := h.dao.listAll(ctx)
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
