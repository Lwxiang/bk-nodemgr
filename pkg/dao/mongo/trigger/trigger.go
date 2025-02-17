/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigger ...
package trigger

import (
	"context"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
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

// upsert updates or inserts a trigger.
func (d *dao) upsert(ctx context.Context, trigger *Trigger) error {
	filter, upsert, opts := buildUpsertParams(trigger)
	result, err := d.client.UpdateOne(ctx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		{
			d.logger.Infof("successfully upserted trigger, unique-key(%s)", trigger.UniqueKey())
		}
	case result.MatchedCount > 0:
		{
			d.logger.Infof("successfully updated trigger, unique-key(%s)", trigger.UniqueKey())
		}
	default:
		d.logger.Warnf("try to upsert trigger but no changes made, unique-key(%s", trigger.UniqueKey())
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(trigger *Trigger) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update trigger by trigger_id.
	filter := bson.D{{Key: "data.trigger_id", Value: trigger.TriggerID}}

	// insert as creation or update data only.
	update := base.BuildUpsertParam(trigger)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// find all trigger.
func (d *dao) find(ctx context.Context, filter bson.D) ([]*Trigger, error) {
	result, err := d.client.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	triggers := make([]*Trigger, 0)
	for result.Next(ctx) {
		table := &TableTrigger{}
		if err := result.Decode(table); err != nil {
			d.logger.Warnf("failed to decode trigger, err %v", err)

			continue
		}
		triggers = append(triggers, table.Data)
	}

	return triggers, nil
}
