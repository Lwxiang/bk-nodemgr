/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operengine ...
package operengine

import (
	"context"
)

// OperationInstStorage represents a OperationInst engine storage handler.
// it will be used to store the custom records during OperationInst scheduling.
type OperationInstStorage interface {
	// CreateOperationInstData will create the OperationInst param.
	CreateOperationInstData(ctx context.Context, data *OperationInstData) error

	// GetOperInstData will get the OperationInst param.
	GetOperInstData(ctx context.Context, operationInstID string) (*OperationInstData, error)

	// UpdateOperationInstData will update the OperationInst param.
	UpdateOperationInstData(ctx context.Context, data *OperationInstData) error

	// MarkOperationInstStopping will mark the OperationInst is stopping.
	MarkOperationInstStopping(ctx context.Context, operationInstID string) error

	// WatchOperInstStopping will return a chan, when the OperationInst is stopping, it will close the chan.
	WatchOperInstStopping(ctx context.Context, operationInstID string) <-chan struct{}
}

// OperationStorage represents a Operation engine storage handler.
type OperationStorage interface {
	// CreateOperation will create an Operation in the database.
	CreateOperation(operation *Operation) error

	// GetOperation will get an Operation from the database.
	GetOperation(operationID string) (*Operation, error)

	// UpdateOperation will update an Operation in the database.
	UpdateOperation(operation *Operation) error
}
