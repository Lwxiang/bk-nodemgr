/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"context"
	"errors"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler ...
type Handler interface {
	// Upsert updates or inserts an OperInstData.
	Upsert(ctx context.Context, data *operengine.OperationInstData) error

	// FindOne ...
	FindOne(ctx context.Context, opts ...OptFn) (*operengine.OperationInstData, error)
}

type handler struct {
	dao *dao
}

// New create a new host handler.
func New(client *mongo.Database, logger logger.Logger) Handler {
	return &handler{
		dao: newDao(client, logger),
	}
}

// Upsert updates or inserts an OperInstData.
func (h *handler) Upsert(ctx context.Context, data *operengine.OperationInstData) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if data == nil {
		return errors.New("data is nil")
	}

	operInstData := &OperInstData{
		OperInstID:        data.OperInstID,
		ActionNames:       data.ActionNames,
		ActionInstDataMap: make(map[string]*ActionInstData, len(data.ActionInstDataMap)),
		OperationDefName:  data.OperationDefName,
		ParentOperInstID:  data.ParentOperInstID,
		Timeout:           data.Timeout,
		InitContent:       data.InitContent,
		CreatedAt:         data.CreatedAt,
		StartedAt:         data.StartedAt,
		EndedAt:           data.EndedAt,
		StoppedAt:         data.StoppedAt,
	}

	for k, v := range data.ActionInstDataMap {
		operInstData.ActionInstDataMap[k] = &ActionInstData{
			OperInstID: v.OperInstID,
			Name:       v.Name,
			Index:      v.Index,
			State:      string(v.State),
			StartedAt:  v.StartedAt,
			EndedAt:    v.EndedAt,
			StoppedAt:  v.StoppedAt,
			Messages:   v.Messages,
			Content:    v.Content,
		}
	}

	err := h.dao.upsert(ctx, operInstData)
	if err != nil {
		return err
	}

	return nil
}

// FindOne ...
func (h *handler) FindOne(ctx context.Context, opts ...OptFn) (*operengine.OperationInstData, error) {
	filter := bson.D{{Key: "basic.is_deleted", Value: false}}
	for _, opt := range opts {
		filter = opt(filter)
	}

	operInstDatas, err := h.dao.find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(operInstDatas) == 0 {
		return nil, errors.New("not found")
	}

	operInstData := operInstDatas[0]

	data := &operengine.OperationInstData{
		OperInstID:        operInstData.OperInstID,
		OperationDefName:  operInstData.OperationDefName,
		ActionNames:       operInstData.ActionNames,
		ActionInstDataMap: make(map[string]*operengine.ActionInstData, len(operInstData.ActionInstDataMap)),
		ParentOperInstID:  operInstData.ParentOperInstID,
		Timeout:           operInstData.Timeout,
		InitContent:       operInstData.InitContent,
		CreatedAt:         operInstData.CreatedAt,
		StartedAt:         operInstData.StartedAt,
		EndedAt:           operInstData.EndedAt,
		StoppedAt:         operInstData.StoppedAt,
	}

	for k, v := range operInstData.ActionInstDataMap {
		data.ActionInstDataMap[k] = &operengine.ActionInstData{
			OperInstID: v.OperInstID,
			Name:       v.Name,
			Index:      v.Index,
			State:      operengine.ActionInstState(v.State),
			StartedAt:  v.StartedAt,
			EndedAt:    v.EndedAt,
			StoppedAt:  v.StoppedAt,
			Messages:   v.Messages,
			Content:    v.Content,
		}
	}

	return data, nil
}
