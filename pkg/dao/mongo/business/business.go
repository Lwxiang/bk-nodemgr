/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package business provides business dao operations.
package business

import (
	"context"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(tenantID string, client *mongo.Database, logger logger.Logger) *dao {
	return &dao{client: client.Collection(TableName(tenantID)), logger: logger}
}

type dao struct {
	client *mongo.Collection
	logger logger.Logger
}

// upsert updates or inserts a business.
func (d *dao) upsert(ctx context.Context, biz *Business) error {
	filter, upsert, opts := buildUpsertParams(biz)
	result, err := d.client.UpdateOne(ctx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		{
			d.logger.Infof("successfully upserted business, unique-key(%s)", biz.UniqueKey())
		}
	case result.MatchedCount > 0:
		{
			d.logger.Infof("successfully updated business, unique-key(%s)", biz.UniqueKey())
		}
	default:
		d.logger.Warnf("try to upsert business but no changes made, unique-key(%s", biz.UniqueKey())
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(biz *Business) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	nowTime := time.Now()

	// update business by biz_id.
	filter := bson.D{{Key: "data.biz_id", Value: biz.BizID}}

	// insert as creation or update data only.
	update := bson.D{
		{
			Key: "$set",
			Value: bson.M{
				"basic.is_deleted": false,
				"basic.updated_at": nowTime,
				"data":             biz,
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

// ListAll list all business.
func (d *dao) listAll(ctx context.Context) ([]*Business, error) {
	result, err := d.client.Find(ctx, bson.D{{Key: "basic.is_deleted", Value: false}})
	if err != nil {
		return nil, err
	}

	bizs := make([]*Business, 0)
	for result.Next(ctx) {
		table := &TableBusiness{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode business, err %v", err)

			continue
		}
		bizs = append(bizs, table.Data)
	}

	return bizs, nil
}
