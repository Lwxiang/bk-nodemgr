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
	// ActionInstanceStatePending action instance state pending.
	ActionInstanceStatePending ActionInstState = "pending"

	// ActionInstanceStateRunning action instance state running.
	ActionInstanceStateRunning ActionInstState = "running"

	// ActionInstanceStateSuccess action instance state success.
	ActionInstanceStateSuccess ActionInstState = "success"

	// ActionInstanceStateFailed action instance state failed.
	ActionInstanceStateFailed ActionInstState = "failed"

	// ActionInstanceStateTimeout action instance state timeout.
	ActionInstanceStateTimeout ActionInstState = "timeout"

	// ActionInstanceStateSkipped action instance state skipped.
	ActionInstanceStateSkipped ActionInstState = "skipped"

	// ActionInstanceStateTerminated action instance state terminated.
	ActionInstanceStateTerminated ActionInstState = "terminated"

	// ActionInstanceStateUnknown action instance state unknown.
	ActionInstanceStateUnknown ActionInstState = "unknown"
)

// OperationInstState OperationInst state.
type OperationInstState string

const (
	// OperationInstStatePending OperationInst state pending.
	OperationInstStatePending OperationInstState = "pending"

	// OperationInstStateRunning OperationInst state running.
	OperationInstStateRunning OperationInstState = "running"

	// OperationInstStateSuccess OperationInst state success.
	OperationInstStateSuccess OperationInstState = "success"

	// OperationInstStateFailed OperationInst state failed.
	OperationInstStateFailed OperationInstState = "failed"

	// OperationInstStateTimeout OperationInst state timeout.
	OperationInstStateTimeout OperationInstState = "timeout"

	// OperationInstStateSkipped OperationInst state skipped.
	OperationInstStateSkipped OperationInstState = "skipped"

	// OperationInstStateTerminated OperationInst state terminated.
	OperationInstStateTerminated OperationInstState = "terminated"
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
