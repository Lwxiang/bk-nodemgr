/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigengine ...
package trigengine

import (
	"context"
	"errors"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/trigger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/trigengine"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName ...
const StorageName = "trigengine"

// NewStorage ...
func NewStorage(client *mongo.Client, database string, logger logger.Logger) (Storage, error) {
	if client == nil {
		return nil, errors.New("mongo client is nil")
	}

	s := &storage{
		Storage: base.Storage{
			Name:     StorageName,
			Database: client.Database(database),
			Logger:   logger,
		},
	}
	err := base.InitStorage(&s.Storage,
		base.WithStartFunc(s.initDao),
		base.WithCheckFunc(s.check))
	if err != nil {
		s.Logger.Errorf("new storage failed, err: %v", err)
		return nil, err
	}

	return s, nil
}

type storage struct {
	base.Storage

	// dao
	triggerDao trigger.Handler
}

func (s *storage) initDao() error {
	s.triggerDao = trigger.New(s.Database, s.Logger)

	return nil
}

func (s *storage) check() error {
	if s.triggerDao == nil {
		return errors.New("trigger dao is nil")
	}

	return nil
}

// Create ...
func (s *storage) Create(ctx context.Context, trigger *trigengine.Trigger) error {
	if err := s.triggerDao.Upsert(ctx, trigger); err != nil {
		s.Logger.Errorf("create failed, err: %v", err)
		return err
	}

	return nil
}

// Get ...
func (s *storage) Get(ctx context.Context, triggerID string) (*trigengine.Trigger, error) {
	one, err := s.triggerDao.FindOne(ctx, trigger.WithTriggerID(triggerID))
	if err != nil {
		s.Logger.Errorf("get one failed, err: %v", err)
		return nil, err
	}

	return one, nil
}

// ListAliveOnceTrigger ...
func (s *storage) ListAliveOnceTrigger(_ context.Context) ([]*trigengine.Trigger, error) {
	// TODO implement me
	panic("implement me")
}

// AllPeriodicTrigger ...
func (s *storage) AllPeriodicTrigger(ctx context.Context) ([]*trigengine.Trigger, error) {
	triggers, err := s.triggerDao.Find(ctx,
		trigger.WithStatus(trigengine.StateInit, trigengine.StateRunning),
		trigger.WithCategory(trigengine.CategoryPeriodic))
	if err != nil {
		s.Logger.Errorf("list alive periodic trigger failed, err: %v", err)

		return nil, err
	}

	return triggers, nil
}

// ListAliveOrderedTrigger ...
func (s *storage) ListAliveOrderedTrigger(_ context.Context) ([]*trigengine.Trigger, error) {
	// TODO implement me
	panic("implement me")
}

// Update ...
func (s *storage) Update(ctx context.Context, trigger *trigengine.Trigger) error {
	err := s.triggerDao.Upsert(ctx, trigger)
	if err != nil {
		s.Logger.Errorf("update failed, err: %v", err)
		return err
	}

	return nil
}
