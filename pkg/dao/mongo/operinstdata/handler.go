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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler ...
type Handler interface {
	// Upsert updates or inserts an OperInstData.
	Upsert(ctx context.Context, data *operengine.OperInstData) error

	// FindOne find one OperInstData.
	FindOne(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error)

	// RefreshActInstDataMsg refreshes the message of an ActionInstData.
	RefreshActInstDataMsg(ctx context.Context, data *operengine.ActionInstData) error
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
func (h *handler) Upsert(ctx context.Context, data *operengine.OperInstData) error {
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
		OperDefName:       data.OperDefName,
		ParentOperInstID:  data.ParentOperInstID,
		Timeout:           data.Timeout,
		State:             string(data.State),
		CreatedAt:         data.CreatedAt,
		StartedAt:         data.StartedAt,
		EndedAt:           data.EndedAt,
		StoppedAt:         data.StoppedAt,
	}

	if data.InitContent == nil {
		return errors.New("invalid init content")
	}

	bytes, err := json.Marshal(data.InitContent)
	if err != nil {
		return err
	}

	operInstData.InitContent = string(bytes)

	for k, v := range data.ActionInstDataMap {
		actionInstData := &ActionInstData{
			TriggerID:  v.TriggerID,
			OperInstID: v.OperInstID,
			Name:       v.Name,
			Index:      v.Index,
			State:      string(v.State),
			StartedAt:  v.StartedAt,
			EndedAt:    v.EndedAt,
			StoppedAt:  v.StoppedAt,
		}

		for _, message := range v.Messages {
			actionInstData.Messages = append(actionInstData.Messages, Message{
				Time: message.Time,
				Text: message.Text,
			})
		}

		bytes, err := json.Marshal(v.Content)
		if err != nil {
			return err
		}

		actionInstData.Content = string(bytes)

		operInstData.ActionInstDataMap[k] = actionInstData
	}

	err = h.dao.upsert(ctx, operInstData)
	if err != nil {
		return err
	}

	return nil
}

// FindOne ...
func (h *handler) FindOne(ctx context.Context, opts ...OptFn) (*operengine.OperInstData, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

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

	data := &operengine.OperInstData{
		OperInstID:        operInstData.OperInstID,
		OperDefName:       operInstData.OperDefName,
		ActionNames:       operInstData.ActionNames,
		ActionInstDataMap: make(map[string]*operengine.ActionInstData, len(operInstData.ActionInstDataMap)),
		ParentOperInstID:  operInstData.ParentOperInstID,
		Timeout:           operInstData.Timeout,
		CreatedAt:         operInstData.CreatedAt,
		StartedAt:         operInstData.StartedAt,
		State:             operengine.OperInstState(operInstData.State),
		EndedAt:           operInstData.EndedAt,
		StoppedAt:         operInstData.StoppedAt,
	}

	if len(operInstData.InitContent) == 0 {
		return nil, errors.New("invalid init content")
	}

	err = json.Unmarshal([]byte(operInstData.InitContent), &data.InitContent)
	if err != nil {
		return nil, err
	}

	for k, v := range operInstData.ActionInstDataMap {
		actionInstData := &operengine.ActionInstData{
			TriggerID:  v.TriggerID,
			OperInstID: v.OperInstID,
			Name:       v.Name,
			Index:      v.Index,
			State:      operengine.ActionInstState(v.State),
			StartedAt:  v.StartedAt,
			EndedAt:    v.EndedAt,
			StoppedAt:  v.StoppedAt,
		}

		for _, msg := range v.Messages {
			actionInstData.Messages = append(actionInstData.Messages, operengine.Message{
				Time: msg.Time,
				Text: msg.Text,
			})
		}

		err = json.Unmarshal([]byte(v.Content), &actionInstData.Content)
		if err != nil {
			return nil, err
		}

		data.ActionInstDataMap[k] = actionInstData
	}

	return data, nil
}

// RefreshActInstDataMsg ...
func (h *handler) RefreshActInstDataMsg(ctx context.Context, data *operengine.ActionInstData) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if data == nil {
		return errors.New("data is nil")
	}

	filter := bson.D{{Key: "basic.is_deleted", Value: false}}
	opts := []OptFn{
		WithTriggerID(data.TriggerID),
		WithOperInstID(data.OperInstID),
	}
	for _, opt := range opts {
		filter = opt(filter)
	}

	filed := fmt.Sprintf("data.action_inst_data_map.%s.messages", data.Name)
	err := h.dao.updateField(ctx, filter, filed, data.Messages)
	if err != nil {
		return err
	}

	return nil
}
