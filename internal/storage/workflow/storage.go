/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package workflow provides implementation of workflow storage.
package workflow

import (
	"context"
	"errors"
	"sync"
	"time"

	baseStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewStorage(config *StorageConfig) *Storage {
	return &Storage{
		config:                 config,
		isRunning:              false,
		stopEventSubscriptions: make(map[string]*StopEventSubscription),
		stoppingTasks:          make(map[string]struct{}),
		checkingC:              make(chan struct{}, 100),
	}
}

type StorageConfig struct {
	MongoDB config.MongoDB

	Database           string
	TaskCollection     string
	StoppingCollection string
}

type StopEventSubscription struct {
	TaskID string
	C      chan<- struct{}
}

type Storage struct {
	// config
	config *StorageConfig

	// state
	isRunning bool

	// mongo
	mongoClient *mongo.Client

	// stop event subscriptions
	stopEventSubscriptions      map[string]*StopEventSubscription
	stopEventSubscriptionsMutex sync.RWMutex

	stoppingTasks      map[string]struct{}
	stoppingTasksMutex sync.RWMutex

	checkingC chan struct{}
}

func (s *Storage) CreateTaskData(data *workflow.TaskData) error {
	_, err := s.mongoClient.Database(s.config.Database).Collection(s.config.TaskCollection).InsertOne(
		context.Background(),
		&TableTaskData{
			BasicInfo: baseStorage.NewBasicInfo(),
			Data:      s.convertTaskData2Table(data),
		})

	return err
}

func (s *Storage) GetTaskData(taskID string) (*workflow.TaskData, error) {
	r := s.mongoClient.Database(s.config.Database).Collection(s.config.TaskCollection).FindOne(context.Background(), bson.M{"data.task_id": taskID})

	table := &TableTaskData{}
	if err := r.Decode(table); err != nil {
		blog.Errorf("failed to decode task data, task-id(%s): %v", taskID, err)

		return nil, err
	}

	return s.convertTable2TaskData(table.Data), nil
}

func (s *Storage) UpdateTaskData(data *workflow.TaskData) error {
	_, err := s.mongoClient.Database(s.config.Database).Collection(s.config.TaskCollection).UpdateOne(context.Background(), bson.M{"data.task_id": data.TaskID}, bson.M{"$set": bson.M{"data": s.convertTaskData2Table(data)}})

	return err
}

func (s *Storage) MarkTaskStopping(taskID string) error {
	_, err := s.mongoClient.Database(s.config.Database).Collection(s.config.StoppingCollection).InsertOne(
		context.Background(),
		&TableStoppingTask{
			BasicInfo: baseStorage.NewBasicInfo(),
			Data: &StoppingTask{
				TaskID:   taskID,
				ExpireAt: time.Now().Add(30 * time.Minute),
			},
		})
	if err != nil {
		blog.Errorf("failed to insert stopping task(%s): %v", taskID, err)

		return err
	}

	blog.Infof("successfully marked task stopping, task(%s)", taskID)

	return nil
}

func (s *Storage) WatchTaskStopping(ctx context.Context, taskID string) <-chan struct{} {
	c := make(chan struct{}, 1)
	subscription := &StopEventSubscription{
		TaskID: taskID,
		C:      c,
	}

	subscriptionID := uuid.New().String()
	go func() {
		<-ctx.Done()

		s.stopEventSubscriptionsMutex.Lock()
		delete(s.stopEventSubscriptions, subscriptionID)
		s.stopEventSubscriptionsMutex.Unlock()
	}()

	s.stopEventSubscriptionsMutex.Lock()
	s.stopEventSubscriptions[subscriptionID] = subscription
	s.stopEventSubscriptionsMutex.Unlock()

	s.checkingC <- struct{}{}

	return c
}

func (s *Storage) Start(ctx context.Context) error {
	if s.isRunning {
		return errors.New("storage already started")
	}

	if err := s.initializeMongoDB(); err != nil {
		return err
	}

	s.isRunning = true
	go s.syncStoppingTasks(ctx)
	go s.watchStoppingTasks(ctx)
	go s.checkNotifyStopping(ctx)

	blog.Info("successfully started storage")
	return nil
}

func (s *Storage) initializeMongoDB() error {
	var err error
	s.mongoClient, err = mongo.NewClient(
		&options.ClientOptions{
			Hosts: s.config.MongoDB.Hosts,
			Auth: &options.Credential{
				Username:      s.config.MongoDB.Username,
				Password:      s.config.MongoDB.Password,
				AuthSource:    s.config.MongoDB.AuthSource,
				AuthMechanism: s.config.MongoDB.AuthMechanism,
			},
		},
	)
	if err != nil {
		blog.Errorf("failed to create mongo client: %v", err)

		return err
	}

	if err = s.mongoClient.Connect(context.Background()); err != nil {
		blog.Errorf("failed to connect mongo client: %v", err)

		return err
	}

	if err = s.mongoClient.Ping(context.Background(), nil); err != nil {
		blog.Errorf("failed to ping mongo client: %v", err)

		return err
	}

	blog.Infof("successfully initialized mongo client: %v", s.config.MongoDB.Hosts)
	return nil
}

func (s *Storage) syncStoppingTasks(ctx context.Context) {
	collection := s.mongoClient.Database(s.config.Database).Collection(s.config.StoppingCollection)
	findOptions := options.Find()
	findOptions.SetLimit(1000)

	for {
		select {
		case <-ctx.Done():
			return

		default:
			cursor, err := collection.Find(context.Background(), bson.D{}, findOptions)
			if err != nil {
				blog.Errorf("failed to find stopping tasks: %v", err)

				break
			}
			defer cursor.Close(context.TODO())

			stoppingTasks := make(map[string]struct{})
			for cursor.Next(context.TODO()) {
				stoppingTask := &StoppingTask{}

				if err = cursor.Decode(stoppingTask); err != nil {
					blog.Errorf("failed to decode stopping task: %v", err)

					continue
				}

				stoppingTasks[stoppingTask.TaskID] = struct{}{}
			}

			if err = cursor.Err(); err != nil {
				blog.Errorf("failed to iterate stopping tasks: %v", err)

				break
			}

			s.stoppingTasksMutex.Lock()
			s.stoppingTasks = stoppingTasks
			s.stoppingTasksMutex.Unlock()

			s.checkingC <- struct{}{}

			time.Sleep(10 * time.Second)
		}
	}
}

func (s *Storage) watchStoppingTasks(ctx context.Context) {
	collection := s.mongoClient.Database(s.config.Database).Collection(s.config.StoppingCollection)

	watchOptions := options.ChangeStream()
	watchOptions.SetBatchSize(1000)

	// only watch insert event
	pipeline := mongo.Pipeline{{{Key: "$match", Value: bson.D{{Key: "operationType", Value: "insert"}}}}}

	for {
		select {
		case <-ctx.Done():
			return

		default:
			changeStream, err := collection.Watch(context.Background(), pipeline, watchOptions)
			if err != nil {
				blog.Errorf("failed to watch stopping tasks, db(%s), collection(%s): %v", s.config.Database, s.config.StoppingCollection, err)

				return
			}

			for changeStream.Next(ctx) {
				var changeEvent TableTaskDataChangeEvent
				if err := changeStream.Decode(&changeEvent); err != nil {
					blog.Errorf("failed to decode change event: %v", err)

					continue
				}

				if changeEvent.FullDocument == nil || changeEvent.FullDocument.Data == nil || changeEvent.FullDocument.Data.TaskID == "" {
					blog.Errorf("failed to get task_id: %v", changeEvent)
					continue
				}

				blog.Infof("got inserted-task-id(%s)", changeEvent.FullDocument.Data.TaskID)

				s.stoppingTasksMutex.Lock()
				s.stoppingTasks[changeEvent.FullDocument.Data.TaskID] = struct{}{}
				s.stoppingTasksMutex.Unlock()

				s.checkingC <- struct{}{}
			}
			changeStream.Close(ctx)

			// backoff several seconds before next watch.
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *Storage) checkNotifyStopping(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-s.checkingC:
			s.stopEventSubscriptionsMutex.Lock()
			for key, subscription := range s.stopEventSubscriptions {
				s.stoppingTasksMutex.Lock()
				_, ok := s.stoppingTasks[subscription.TaskID]
				s.stoppingTasksMutex.Unlock()

				// this task is stopping, notify the subscription
				if ok {
					subscription.C <- struct{}{}
					delete(s.stopEventSubscriptions, key)
				}
			}
			s.stopEventSubscriptionsMutex.Unlock()
		}
	}
}

func (s *Storage) convertTaskData2Table(data *workflow.TaskData) *TaskData {
	actionData := make(map[string]*ActionData)
	for k, v := range data.ActionData {
		actionData[k] = &ActionData{
			Name:      v.Name,
			State:     string(v.State),
			StartedAt: v.StartedAt,
			EndedAt:   v.EndedAt,
			StoppedAt: v.StoppedAt,
			Messages:  v.Messages,
			Content:   v.Content,
		}
	}

	return &TaskData{
		TaskID:       data.TaskID,
		Pipeline:     data.Pipeline,
		Actions:      data.Actions,
		ActionData:   actionData,
		ParentTaskID: data.ParentTaskID,
		Timeout:      data.Timeout,
		IsPeriod:     data.IsPeriod,
		PeriodSpec:   data.PeriodSpec,
		InitContent:  data.InitContent,
		CreatedAt:    data.CreatedAt,
		StartedAt:    data.StartedAt,
		EndedAt:      data.EndedAt,
		StoppedAt:    data.StoppedAt,
	}
}

func (s *Storage) convertTable2TaskData(table *TaskData) *workflow.TaskData {
	actionData := make(map[string]*workflow.ActionData)
	for k, v := range table.ActionData {
		actionData[k] = &workflow.ActionData{
			Name:      v.Name,
			State:     workflow.ActionState(v.State),
			StartedAt: v.StartedAt,
			EndedAt:   v.EndedAt,
			StoppedAt: v.StoppedAt,
			Messages:  v.Messages,
			Content:   v.Content,
		}
	}

	return &workflow.TaskData{
		TaskID:       table.TaskID,
		Pipeline:     table.Pipeline,
		Actions:      table.Actions,
		ActionData:   actionData,
		ParentTaskID: table.ParentTaskID,
		Timeout:      table.Timeout,
		IsPeriod:     table.IsPeriod,
		PeriodSpec:   table.PeriodSpec,
		InitContent:  table.InitContent,
		CreatedAt:    table.CreatedAt,
		StartedAt:    table.StartedAt,
		EndedAt:      table.EndedAt,
		StoppedAt:    table.StoppedAt,
	}
}
