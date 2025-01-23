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
	"errors"

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/storage/base"
	"git.woa.com/bk-gse/bk-nodeman/pkg/dao/mongo/operation"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
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

	daoOperation operation.Handler
}

func (s *storage) initDao() error {
	s.daoOperation = operation.New(s.Database, s.Logger)

	return nil
}

func (s *storage) check() error {
	if s.daoOperation == nil {
		return errors.New("dao operation is nil")
	}

	return nil
}

// CreateOperation ...
func (s *storage) CreateOperation(operation *operengine.Operation) error {
	//TODO implement me
	panic("implement me")
}

// GetOperation ...
func (s *storage) GetOperation(operationID string) (*operengine.Operation, error) {
	//TODO implement me
	panic("implement me")
}

// UpdateOperation ...
func (s *storage) UpdateOperation(operation *operengine.Operation) error {
	//TODO implement me
	panic("implement me")
}
