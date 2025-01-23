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
	"fmt"
	"sync"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/storage/base"
	"git.woa.com/bk-gse/bk-nodeman/pkg/dao/mongo/operinstdata"
	"git.woa.com/bk-gse/bk-nodeman/pkg/dao/mongo/stopoperinst"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/scheduler"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/singleflight"
)

// StorageName ...
const StorageName = "operinstdata"

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
		base.WithStartFunc(s.startFn),
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
	operinstdataDao operinstdata.Handler
	stopoperinstDao stopoperinst.Handler

	// stop event subscriptions
	stopEventSubsMap      map[string]*StopEventSubscription
	stopEventSubsMapMutex sync.RWMutex

	stopOperInsts      map[string]struct{}
	stopOperInstsMutex sync.RWMutex

	sg singleflight.Group
}

func (s *storage) startFn() error {
	s.operinstdataDao = operinstdata.New(s.Database, s.Logger)
	s.stopoperinstDao = stopoperinst.New(s.Database, s.Logger)

	s.stopEventSubsMap = make(map[string]*StopEventSubscription)
	s.stopOperInsts = make(map[string]struct{})

	s.registerScheduler()

	return nil
}

func (s *storage) check() error {
	if s.operinstdataDao == nil {
		return errors.New("operation instance dao is nil")
	}

	return nil
}

func (s *storage) registerScheduler() {
	s.Scheduler = scheduler.NewScheduler(scheduler.WithLogger(s.Logger), scheduler.WithInterval(time.Second*5))
	s.Scheduler.RegisterTask(&scheduler.Task{
		ID:       "sync stopping operation inst",
		Interval: 10 * time.Second,
		Timeout:  20 * time.Second,
		Fn:       s.syncStopOperInsts,
	})

	go s.stopoperinstDao.WatchInsert(func(stopInstID string) {
		s.stopOperInstsMutex.Lock()
		defer s.stopOperInstsMutex.Unlock()
		s.stopOperInsts[stopInstID] = struct{}{}

		go s.checkNotifyStopping(s.Ctx)
	})
}

// StopEventSubscription represents the stop event subscription.
type StopEventSubscription struct {
	OperInstID string
	C          chan<- struct{}
}

// CreateOperationInstData create a new task data.
func (s *storage) CreateOperationInstData(ctx context.Context, data *operengine.OperationInstData) error {
	if ctx == nil {
		return errors.New("context is nil")
	}

	if data == nil {
		return errors.New("operation inst data is nil")
	}

	if err := s.operinstdataDao.Upsert(ctx, data); err != nil {
		return fmt.Errorf("failed to create operation inst data: %v", err)
	}

	return nil
}

// GetOperInstData get task data.
func (s *storage) GetOperInstData(ctx context.Context, operInstID string) (*operengine.OperationInstData, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}

	if operInstID == "" {
		return nil, errors.New("operation inst id is empty")
	}

	data, err := s.operinstdataDao.FindOne(ctx, operinstdata.WithOperInstID(operInstID))
	if err != nil {
		return nil, fmt.Errorf("failed to get operation inst data: %v", err)
	}

	return data, nil
}

// UpdateOperationInstData update task data.
func (s *storage) UpdateOperationInstData(ctx context.Context, data *operengine.OperationInstData) error {
	if err := s.operinstdataDao.Upsert(ctx, data); err != nil {
		return fmt.Errorf("failed to update operation inst data, operation-inst(%v), err: %v", data, err)
	}

	return nil
}

// MarkOperationInstStopping mark task stopping.
func (s *storage) MarkOperationInstStopping(ctx context.Context, operationInstID string) error {
	err := s.stopoperinstDao.Upsert(ctx, operationInstID)
	if err != nil {
		return fmt.Errorf("failed to mark operation inst stopping failed, operation-inst-id(%v), err: %v",
			operationInstID, err)
	}

	return nil
}

// WatchOperInstStopping watch operation instance stopping event.
func (s *storage) WatchOperInstStopping(ctx context.Context, operInstID string) <-chan struct{} {
	c := make(chan struct{}, 1)
	subscription := &StopEventSubscription{
		OperInstID: operInstID,
		C:          c,
	}

	subscriptionID := uuid.New().String()
	s.stopEventSubsMapMutex.Lock()
	s.stopEventSubsMap[subscriptionID] = subscription
	s.stopEventSubsMapMutex.Unlock()

	go func() {
		<-ctx.Done()

		s.stopEventSubsMapMutex.Lock()
		delete(s.stopEventSubsMap, subscriptionID)
		s.stopEventSubsMapMutex.Unlock()
	}()

	go func() {
		err := s.checkNotifyStopping(ctx)
		if err != nil {
			s.Logger.Errorf("watch operation instance stopping event succeed, "+
				"but check notify stopping failed, err: %v", err)
		}
	}()

	return c
}

// syncStopOperInsts sync all stopping operation instances.
func (s *storage) syncStopOperInsts(ctx context.Context) error {
	stopInstIDs, err := s.stopoperinstDao.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to find all stopping operation instances: %v", err)
	}

	stopInstMap := make(map[string]struct{}, len(stopInstIDs))
	for _, stopInstID := range stopInstIDs {
		stopInstMap[stopInstID] = struct{}{}
	}

	s.stopOperInstsMutex.Lock()
	s.stopOperInsts = stopInstMap
	s.stopOperInstsMutex.Unlock()

	go func() {
		err := s.checkNotifyStopping(ctx)
		if err != nil {
			s.Logger.Errorf("sync stopping event succeed, but check notify stopping failed, err: %v", err)
		}
	}()

	return nil
}

// checkNotifyStopping check and notify the stopping event.
func (s *storage) checkNotifyStopping(ctx context.Context) error {
	_, err, _ := s.sg.Do("checkNotifyStopping", func() (interface{}, error) {
		err := s.processStoppingEvents(ctx)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// TODO: 此处有坑，需要重新测试
func (s *storage) processStoppingEvents(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	notifications := s.getNotifications()

	for _, notify := range notifications {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case notify.Subscription.C <- struct{}{}:
			s.removeSubscription(notify.Key)
		default:
			s.Logger.Errorf("failed to notify stopping event, the channel is full, notify: %+v", notify)
		}
	}

	return nil
}

type notifyItem struct {
	Key          string
	Subscription *StopEventSubscription
}

// getNotifications get the notifications.
func (s *storage) getNotifications() []notifyItem {
	s.stopEventSubsMapMutex.Lock()
	defer s.stopEventSubsMapMutex.Unlock()
	s.stopOperInstsMutex.Lock()
	defer s.stopOperInstsMutex.Unlock()

	notifications := make([]notifyItem, 0, len(s.stopEventSubsMap))
	for key, subscription := range s.stopEventSubsMap {
		_, ok := s.stopOperInsts[subscription.OperInstID]
		if ok {
			notifications = append(notifications, notifyItem{
				Key:          key,
				Subscription: subscription,
			})
		}
	}

	return notifications
}

func (s *storage) removeSubscription(key string) {
	s.stopEventSubsMapMutex.Lock()
	defer s.stopEventSubsMapMutex.Unlock()

	delete(s.stopEventSubsMap, key)
}
