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

	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Storage defines the storage interface.
type Storage interface {
	// UpsertBusiness updates or inserts a business.
	UpsertBusiness(biz *types.Business) error
}

// NewStorage creates a new topology storage.
func NewStorage(config *StorageConfig) *DefaultStorage {
	return &DefaultStorage{
		config:    config,
		isRunning: false,
	}
}

// StorageConfig defines the config of the topology storage.
type StorageConfig struct {
	MongoDB config.MongoDB

	Database string

	BusinessCollection string
	HostCollection     string
}

// DefaultStorage implements the Storage interface.
type DefaultStorage struct {
	// config
	config *StorageConfig

	// state
	isRunning bool

	// ctx
	ctx    context.Context
	cancel context.CancelFunc

	// mongo
	mongoClient *mongo.Client
}

// Start starts the topology storage.
func (ds *DefaultStorage) Start(ctx context.Context) error {
	if ds.isRunning {
		return errors.New("storage already started")
	}

	if err := ds.initializeMongoDB(ctx); err != nil {
		return err
	}

	ds.ctx, ds.cancel = context.WithCancel(ctx)
	ds.isRunning = true

	blog.Info("successfully started storage")

	return nil
}

func (ds *DefaultStorage) initializeMongoDB(ctx context.Context) error {
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
		blog.Errorf("failed to connect to mongo client: %v", err)

		return err
	}

	if err = ds.mongoClient.Ping(ctx, nil); err != nil {
		blog.Errorf("failed to ping mongo client: %v", err)

		return err
	}

	blog.Infof("successfully initialized mongo client: %v", ds.config.MongoDB.Hosts)

	return nil
}

// UpsertBusiness updates or inserts a business.
func (ds *DefaultStorage) UpsertBusiness(biz *types.Business) error {
	data := &Business{}
	data.fromRuntime(biz)

	filter, upsert, opts := (&TableBusiness{}).upsertParams(data)

	result, err := ds.mongoClient.Database(ds.config.Database).Collection(ds.config.BusinessCollection).
		UpdateOne(ds.ctx, filter, upsert, opts)
	if err != nil {
		return err
	}

	if result.UpsertedCount > 0 {
		blog.Infof("successfully upserted business(%d)", data.BizID)

		return nil
	}

	if result.MatchedCount > 0 {
		blog.Infof("successfully updated business(%d)", data.BizID)

		return nil
	}

	blog.Warnf("try to upsert business but no changes made. biz-id(%d)", data.BizID)

	return nil
}

// ListBusinesses lists all businesses.
func (ds *DefaultStorage) ListBusinesses() ([]*types.Business, error) {
	data := make([]*types.Business, 0)

	result, err := ds.mongoClient.Database(ds.config.Database).Collection(ds.config.BusinessCollection).
		Find(ds.ctx, bson.D{{Key: "basic.is_deleted", Value: false}})
	if err != nil {
		return nil, err
	}

	for result.Next(ds.ctx) {
		table := &TableBusiness{}
		if err := result.Decode(table); err != nil {
			blog.Warnf("failed to decode business: %v", err)

			continue
		}
		data = append(data, table.Data.toRuntime())
	}

	return data, nil
}
