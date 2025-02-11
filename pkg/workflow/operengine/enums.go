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

// ActionInstState action instance state.
type ActionInstState string

const (
	// ActionInstStatePending action instance state pending.
	ActionInstStatePending ActionInstState = "pending"

	// ActionInstStateRunning action instance state running.
	ActionInstStateRunning ActionInstState = "running"

	// ActionInstStateSuccess action instance state success.
	ActionInstStateSuccess ActionInstState = "success"

	// ActionInstStateFailed action instance state failed.
	ActionInstStateFailed ActionInstState = "failed"

	// ActionInstStateTimeout action instance state timeout.
	ActionInstStateTimeout ActionInstState = "timeout"

	// ActionInstStateSkipped action instance state skipped.
	ActionInstStateSkipped ActionInstState = "skipped"

	// ActionInstStateTerminated action instance state terminated.
	ActionInstStateTerminated ActionInstState = "terminated"

	// ActionInstanceStateUnknown action instance state unknown.
	ActionInstanceStateUnknown ActionInstState = "unknown"
)

// OperInstState OperInst State.
type OperInstState string

const (
	// OperInstStatePending OperInst state pending.
	OperInstStatePending OperInstState = "pending"

	// OperInstStateRunning OperInst state running.
	OperInstStateRunning OperInstState = "running"

	// OperInstStateSuccess OperInst state success.
	OperInstStateSuccess OperInstState = "success"

	// OperInstStateFailed OperInst state failed.
	OperInstStateFailed OperInstState = "failed"

	// OperInstStateTimeout OperInst state timeout.
	OperInstStateTimeout OperInstState = "timeout"

	// OperInstStateSkipped OperInst state skipped.
	OperInstStateSkipped OperInstState = "skipped"

	// OperInstStateTerminated OperInst state terminated.
	OperInstStateTerminated OperInstState = "terminated"
)

// OperationState represents the state of an operation.
type OperationState string

const (
	// OperationStateInit Operation state init represents the initial state of a operation.
	OperationStateInit OperationState = "init"

	// OperationStateRunning Operation state running represents a operation that is currently running.
	OperationStateRunning OperationState = "running"

	// OperationStateSuccess Operation state success represents a operation that has completed successfully.
	OperationStateSuccess OperationState = "success"

	// OperationStateFailed Operation state failed represents a operation that has failed.
	OperationStateFailed OperationState = "failed"

	// OperationStateTerminated Operation state terminated represents a operation that has been terminated.
	OperationStateTerminated OperationState = "terminated"
)
