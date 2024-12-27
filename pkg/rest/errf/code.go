/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package errf provides blueking error code.
package errf

// NOTE: 错误码规则
// 38号段 + 5位错误码共7位
// 注意：
// - 特殊错误码, 3830403（未授权）, 内部保留

// common error code.
const (
	// OK means success.
	OK int = 0
	// PermissionDenied means the user has no permission.
	PermissionDenied int = 3830403
)

// SysAlarmFlag is the system alarm flag.
const SysAlarmFlag = "SysAlarmFlag"

// Note:
// this scope's error code ranges at [4000000, 4089999], and works for all the scenario
// except sidecar related scenario.
const (
	// Unknown is unknown error, it is always used when an
	// error is wrapped, but the error code is not parsed.
	Unknown int = 3800000
	// InvalidParameter means the request parameter is invalid.
	InvalidParameter int = 3800001
	// TooManyRequest means the incoming request have already exceeded the max limit.
	// and the incoming request is rejected.
	TooManyRequest int = 3800002
	// RecordNotFound means resource not exist.
	RecordNotFound int = 3800003
	// DecodeRequestFailed means decode the request body failed.
	DecodeRequestFailed int = 3800004
	// UnHealthy means service health check failed, current service is not healthy.
	UnHealthy int = 3800005
	// Aborted means the request is aborted because of some unexpected exceptions.
	Aborted int = 3800006
	// DoAuthorizeFailed try to do user's operate authorize, but got an error,
	// so we do not know if the user has the permission or not.
	DoAuthorizeFailed int = 3800007
	// PartialFailed means batch operation is partially failed.
	PartialFailed int = 3800008
	// DBExecCmdFailed means exec database command failed.
	DBExecCmdFailed int = 3800009
	// InvalidCache means cache is invalid.
	InvalidCache int = 3800010

	// MaxErrCode this code is used to check if the error code is out of range.
	MaxErrCode int = 3900000
)
