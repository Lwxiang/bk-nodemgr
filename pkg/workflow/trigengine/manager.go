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
	"fmt"
	"sync"
	"time"

	"github.com/RichardKnop/machinery/v2"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/scheduler"
)

// Manager ...
type Manager struct {
	server  *machinery.Server
	storage Storage

	scheduler   scheduler.Scheduler
	logger      logger.Logger
	lockFactory locker.MutexFactory

	onceTList              []*Trigger
	onceTListMutex         sync.Mutex
	onceTListSyncInterval  time.Duration
	onceTListCheckInterval time.Duration

	orderedTList              []*Trigger
	orderedTListMutex         sync.Mutex
	orderedTListSyncInterval  time.Duration
	orderedTListCheckInterval time.Duration

	periodicTList              []*Trigger
	periodicTListMutex         sync.Mutex
	periodicTListSyncInterval  time.Duration
	periodicTListCheckInterval time.Duration
}

// NewOnceTrigger create a new once trigger.
func (m *Manager) NewOnceTrigger(ctx context.Context, metadata MetadataPeriodic) (string, error) {
	trigger := NewTrigger(CategoryOnce, metadata)
	err := m.storage.Create(ctx, trigger)
	if err != nil {
		return "", err
	}

	defer func() {
		if storeErr := m.storage.Update(ctx, trigger); storeErr != nil {
			err = fmt.Errorf("save trigger failed, original-err: %s, store-err: %s", err, storeErr)
		}
	}()

	if err = trigger.Execute(ctx); err != nil {
		return "", err
	}

	if err = trigger.Terminate(); err != nil {
		return "", err
	}

	return trigger.TriggerID, nil
}

const (
	defaultTimeout = 1 * time.Minute

	// onceTListSyncSpecDefault means sync once trigger every 5 seconds.
	onceTListSyncIntervalDefault = time.Second

	// onceTListCheckIntervalDefault means check once trigger every 5 seconds.
	onceTListCheckIntervalDefault = time.Second

	// orderedTListSyncIntervalDefault means sync ordered trigger every 5 seconds.
	orderedTListSyncIntervalDefault = time.Second

	// orderedTListCheckIntervalDefault means check ordered trigger every 5 seconds.
	orderedTListCheckIntervalDefault = time.Second

	// checkTriggerSpecPeriodic means check periodic trigger every second.
	periodicTListSyncIntervalDefault = 5 * time.Second

	// syncTriggerSpecPeriodic means sync periodic trigger every second.
	periodicTListCheckIntervalDefault = 5 * time.Second
)

// OptionFunc ...
type OptionFunc func(*Manager)

// WithLogger ...
func WithLogger(logger logger.Logger) OptionFunc {
	return func(m *Manager) {
		m.logger = logger
	}
}

const (
	taskIDSyncOnceTrigger      = "sync_once_trigger"
	taskIDSyncPeriodicTrigger  = "sync_periodic_trigger"
	taskIDSyncOrderedTrigger   = "sync_ordered_trigger"
	taskIDCheckOnceTrigger     = "check_once_trigger"
	taskIDCheckPeriodicTrigger = "check_periodic_trigger"
	taskIDCheckOrderedTrigger  = "check_ordered_trigger"
)

// NewManager ...
func NewManager(server *machinery.Server, storage Storage, lfactory locker.MutexFactory, opts ...OptionFunc) *Manager {
	m := &Manager{
		server:                     server,
		storage:                    storage,
		logger:                     logger.LoggerDefault{},
		lockFactory:                lfactory,
		onceTList:                  make([]*Trigger, 0),
		onceTListMutex:             sync.Mutex{},
		onceTListSyncInterval:      onceTListSyncIntervalDefault,
		onceTListCheckInterval:     onceTListCheckIntervalDefault,
		orderedTList:               make([]*Trigger, 0),
		orderedTListMutex:          sync.Mutex{},
		orderedTListSyncInterval:   orderedTListSyncIntervalDefault,
		orderedTListCheckInterval:  orderedTListCheckIntervalDefault,
		periodicTList:              make([]*Trigger, 0),
		periodicTListMutex:         sync.Mutex{},
		periodicTListSyncInterval:  periodicTListSyncIntervalDefault,
		periodicTListCheckInterval: periodicTListCheckIntervalDefault,
	}

	for _, fn := range opts {
		fn(m)
	}

	m.initScheduler()

	m.scheduler.Start()

	return m
}

// initScheduler ...
func (m *Manager) initScheduler() {
	m.scheduler = scheduler.NewScheduler(scheduler.WithLogger(m.logger))
	m.registerSyncTask()
	m.registerCheckTask()
}

func (m *Manager) registerSyncTask() {
	m.scheduler.RegisterTask(&scheduler.Task{
		ID:       taskIDSyncOnceTrigger,
		Interval: m.onceTListSyncInterval,
		Timeout:  defaultTimeout,
		Fn: func(ctx context.Context) error {
			err := m.syncOnceTrigger(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	})

	m.scheduler.RegisterTask(&scheduler.Task{
		ID:       taskIDSyncPeriodicTrigger,
		Interval: m.periodicTListSyncInterval,
		Timeout:  defaultTimeout,
		Fn: func(ctx context.Context) error {
			err := m.syncPeriodicTrigger(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	})

	m.scheduler.RegisterTask(&scheduler.Task{
		ID:       taskIDSyncOrderedTrigger,
		Interval: m.orderedTListSyncInterval,
		Timeout:  defaultTimeout,
		Fn: func(ctx context.Context) error {
			err := m.syncOrderedTrigger(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	})
}

func (m *Manager) registerCheckTask() {
	m.scheduler.RegisterTask(&scheduler.Task{
		ID:       taskIDCheckOnceTrigger,
		Interval: m.onceTListCheckInterval,
		Timeout:  defaultTimeout,
		Fn: func(ctx context.Context) error {
			m.onceTListMutex.Lock()
			list := m.onceTList
			m.onceTListMutex.Unlock()

			err := m.checkTriggerList(ctx, list)
			if err != nil {
				return err
			}

			return nil
		},
	})

	m.scheduler.RegisterTask(&scheduler.Task{
		ID:       taskIDCheckPeriodicTrigger,
		Interval: m.periodicTListCheckInterval,
		Timeout:  defaultTimeout,
		Fn: func(ctx context.Context) error {
			m.periodicTListMutex.Lock()
			list := m.periodicTList
			m.periodicTListMutex.Unlock()

			err := m.checkTriggerList(ctx, list)
			if err != nil {
				return err
			}

			return nil
		},
	})

	m.scheduler.RegisterTask(&scheduler.Task{
		ID:       taskIDCheckOrderedTrigger,
		Interval: m.orderedTListCheckInterval,
		Timeout:  defaultTimeout,
		Fn: func(ctx context.Context) error {
			m.orderedTListMutex.Lock()
			list := m.orderedTList
			m.orderedTListMutex.Unlock()

			err := m.checkTriggerList(ctx, list)
			if err != nil {
				return err
			}

			return nil
		},
	})
}

// syncOnceTrigger ...
func (m *Manager) syncOnceTrigger(ctx context.Context) error {
	list, err := m.storage.ListAliveOnceTrigger(ctx)
	if err != nil {
		return err
	}

	m.onceTListMutex.Lock()
	defer m.onceTListMutex.Unlock()

	m.onceTList = list

	return nil
}

// syncPeriodicTrigger ...
func (m *Manager) syncPeriodicTrigger(ctx context.Context) error {
	list, err := m.storage.AllPeriodicTrigger(ctx)
	if err != nil {
		return err
	}

	m.periodicTListMutex.Lock()
	defer m.periodicTListMutex.Unlock()

	m.periodicTList = list

	return nil
}

// syncOrderedTrigger ...
func (m *Manager) syncOrderedTrigger(ctx context.Context) error {
	list, err := m.storage.ListAliveOrderedTrigger(ctx)
	if err != nil {
		return err
	}

	m.orderedTListMutex.Lock()
	defer m.orderedTListMutex.Unlock()

	m.orderedTList = list

	return nil
}

// defaultCheckConcurrency ...
const defaultCheckConcurrency = 100

// checkTriggerList ...
func (m *Manager) checkTriggerList(ctx context.Context, list []*Trigger) error {
	gp := gopool.NewPool()
	gp.SetLimit(defaultCheckConcurrency)

	for idx := range list {
		trigger := list[idx]
		fn := func() error {
			mutex := m.lockFactory.NewMutex(trigger.TriggerID)

			err := mutex.TryLock()
			if err != nil {
				m.logger.Errorf("trigger lockFactory failed, trigger-id:(%s), err: %v", trigger.TriggerID, err)
				return nil
			}
			defer mutex.Unlock()

			if trigger.CheckFeasibility(ctx) != nil {
				return nil
			}

			err = trigger.Execute(ctx)
			if err != nil {
				// to check all triggers no error is returned.
				m.logger.Errorf("trigger execute failed, trigger-id:(%s), err: %v", trigger.TriggerID, err)

				return nil
			}

			return nil
		}

		gp.Go(fn)
	}

	if err := gp.Wait(); err != nil {
		m.logger.Errorf("check periodic trigger failed, err: %v", err)
		return err
	}

	return nil
}
