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

import "context"

// Storage represents a Trigger engine storage handler.
type Storage interface {
	// Create will create a Trigger in the database.
	Create(ctx context.Context, trigger *Trigger) error

	// Get will get a Trigger from the database.
	Get(ctx context.Context, triggerID string) (*Trigger, error)

	// ListAliveOnceTrigger will get all alive once Trigger from the database.
	ListAliveOnceTrigger(ctx context.Context) ([]*Trigger, error)

	// AllPeriodicTrigger will get all periodic Trigger from the database.
	AllPeriodicTrigger(ctx context.Context) ([]*Trigger, error)

	// ListAliveOrderedTrigger will get all alive ordered Trigger from the database.
	ListAliveOrderedTrigger(ctx context.Context) ([]*Trigger, error)

	// Update will update a Trigger in the database.
	Update(ctx context.Context, trigger *Trigger) error
}
