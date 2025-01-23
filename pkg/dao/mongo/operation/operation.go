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
	"context"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, logger logger.Logger) *dao {
	return &dao{client: client.Collection(TableName), logger: logger}
}

type dao struct {
	client *mongo.Collection
	logger logger.Logger
}

// upsert updates or inserts an operation.
func (d *dao) upsert(ctx context.Context, operation *Operation) error {
	filter, upsert, opts := buildUpsertParams(operation)
	result, err := d.client.UpdateOne(ctx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		{
			d.logger.Infof("successfully upserted operation, unique-key(%s)", operation.UniqueKey())
		}
	case result.MatchedCount > 0:
		{
			d.logger.Infof("successfully updated operation, unique-key(%s)", operation.UniqueKey())
		}
	default:
		d.logger.Warnf("try to upsert operation but no changes made, unique-key(%s", operation.UniqueKey())
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(operation *Operation) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	nowTime := time.Now()

	// update operation by operation_id
	filter := bson.D{{Key: "data.operation_id", Value: operation.OperationID}}

	// insert as creation or update data only.
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				"basic.is_deleted": false,
				"basic.updated_at": nowTime,
				"data":             operation,
			},
		},
		{
			Key: "$setOnInsert",
			Value: bson.M{
				"basic.created_at": nowTime,
			},
		},
	}

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// find all operations.
func (d *dao) find(ctx context.Context, filter bson.D) ([]*Operation, error) {
	result, err := d.client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	operations := make([]*Operation, 0)
	for result.Next(ctx) {
		table := &TableOperation{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode trigger, err %v", err)

			continue
		}
		operations = append(operations, table.Data)
	}

	return operations, nil
}
