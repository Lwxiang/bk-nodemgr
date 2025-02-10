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

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/storage/base"
	"git.woa.com/bk-gse/bk-nodeman/pkg/dao/mongo/business"
	"git.woa.com/bk-gse/bk-nodeman/pkg/dao/mongo/host"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// StorageName ...
const StorageName = "topo"

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

// storage implements the Storage interface.
type storage struct {
	base.Storage

	daoBusiness business.Handler

	daoHost host.Handler
}

func (ds *storage) initDao() error {
	ds.daoBusiness = business.New(ds.Database, ds.Logger)
	ds.daoHost = host.New(ds.Database, ds.Logger)

	return nil
}

func (ds *storage) check() error {
	if ds.daoBusiness == nil {
		return errors.New("dao business is nil")
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

// UpsertHosts ...
func (ds *storage) UpsertHosts(ctx context.Context, hosts ...*types.Host) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	if err := ds.daoHost.UpsertMany(ctx, hosts); err != nil {
		return fmt.Errorf("failed to upsert hosts: %v", err)
	}

	return nil
}
