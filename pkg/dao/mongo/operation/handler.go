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
	"context"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/mongo"
	workflows "google.golang.org/api/workflows/v1beta"
)

// Handler operation handler interface.
type Handler interface {
	Create(ctx context.Context, operation *workflows.Operation) error
	Update(ctx context.Context, operation *workflows.Operation) error
	Upsert(ctx context.Context, operation *workflows.Operation) error
	FindAll(ctx context.Context) (*workflows.Operation, error)
	Find(ctx context.Context, opts ...OptFn) (*workflows.Operation, error)
	FindOne(ctx context.Context, opts ...OptFn) (*workflows.Operation, error)
}

// New create a new business handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		dao: newDao(client, logger),
	}
}

type handler struct {
	dao *dao
}

func (h *handler) Create(ctx context.Context, operation *workflows.Operation) error {
	//TODO implement me
	panic("implement me")
}

func (h *handler) Update(ctx context.Context, operation *workflows.Operation) error {
	//TODO implement me
	panic("implement me")
}

func (h *handler) FindAll(ctx context.Context) (*workflows.Operation, error) {
	//TODO implement me
	panic("implement me")
}

func (h *handler) Find(ctx context.Context, opts ...OptFn) (*workflows.Operation, error) {
	//TODO implement me
	panic("implement me")
}

func (h *handler) FindOne(ctx context.Context, opts ...OptFn) (*workflows.Operation, error) {
	//TODO implement me
	panic("implement me")
}

// Upsert ...
func (h *handler) Upsert(ctx context.Context, operation *workflows.Operation) error {
	//TODO implement me
	panic("implement me")
}
