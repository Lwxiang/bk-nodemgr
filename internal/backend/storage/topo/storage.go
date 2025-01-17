/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package topo provides topology storage for nodeman.
package topo

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"git.woa.com/bk-gse/bk-nodeman/pkg/dao/mongo/business"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	pingTimeoutDefault = 3 * time.Second
)

// NewStorage creates a new topology storage.
func NewStorage(config *Config, logger logger.Logger) Storage {
	return &storage{
		config:    config,
		isRunning: false,
		logger:    logger,
	}
}

// Config defines the config of the topology storage.
type Config struct {
	MongoDB config.MongoDB

	Database string
}

// storage implements the Storage interface.
type storage struct {
	// config
	config *Config

	// state
	isRunning bool

	// ctx
	ctx    context.Context
	cancel context.CancelFunc

	logger logger.Logger

	// mongo
	mongoClient   *mongo.Client
	mongoDatabase *mongo.Database

	daoBusiness business.Handler
}

// Start starts the topology storage.
func (ds *storage) Start(ctx context.Context) error {
	if ds.isRunning {
		ds.logger.Warn("trying to start storage, but it is already running, stack(%v)", debug.Stack())
		return errors.New("storage already started")
	}

	if err := ds.initMongoDB(ctx); err != nil {
		ds.logger.Errorf("failed to init mongo client, err: %v", err)
		return err
	}

	ds.initDao()

	ds.ctx, ds.cancel = context.WithCancel(ctx)
	ds.isRunning = true

	ds.logger.Info("successfully started storage")

	return nil
}

func (ds *storage) initMongoDB(ctx context.Context) error {
	var err error
	ds.mongoClient, err = mongo.Connect(
		ctx,
		&options.ClientOptions{
			Hosts: ds.config.MongoDB.Hosts,
			Auth: &options.Credential{
				Username:      ds.config.MongoDB.Username,
				Password:      ds.config.MongoDB.Password,
				AuthSource:    ds.config.MongoDB.AuthSource,
				AuthMechanism: ds.config.MongoDB.AuthMechanism,
			},
		},
	)
	if err != nil {
		ds.logger.Errorf("failed to connect to mongo client, err: %v", err)

		return err
	}

	ds.mongoDatabase = ds.mongoClient.Database(ds.config.Database)

	if err = ds.mongoClient.Ping(ctx, nil); err != nil {
		ds.logger.Errorf("failed to ping mongo client, err: %v", err)

		return err
	}

	ds.logger.Infof("successfully initialized mongo client, hosts(%v)", ds.config.MongoDB.Hosts)

	return nil
}

func (ds *storage) initDao() {
	ds.daoBusiness = business.New(ds.mongoDatabase, ds.logger)
}

// CheckHealthz checks the healthz of the topology storage.
func (ds *storage) CheckHealthz() error {
	if ds.mongoDatabase == nil {
		return errors.New("mongo client not initialized")
	}

	ctx, cancel := context.WithDeadline(ds.ctx, time.Now().Add(pingTimeoutDefault))
	defer cancel()

	if err := ds.mongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping mongo client: %v", err)
	}

	return nil
}

// UpsertBusiness updates or inserts a business.
func (ds *storage) UpsertBusiness(ctx context.Context, biz *types.Business) error {
	if err := ds.daoBusiness.Upsert(ctx, biz); err != nil {
		return fmt.Errorf("failed to upsert business: %v", err)
	}

	return nil
}

// ListBusinesses lists all businesses.
func (ds *storage) ListBusinesses(ctx context.Context) ([]*types.Business, error) {
	bizs, err := ds.daoBusiness.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return bizs, nil
}
