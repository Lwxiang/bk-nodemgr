/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package types define all common types used in nodeman runtime.
// Everything from API or Database should be converted into types in this package before using.
package types

// Tenant represents a blueking tenant.
type Tenant struct {
	// tenant-id is the unique identifier for a tenant in a Blueking environment.
	TenantID string

	// there is only one admin tenant in a Blueking environment.
	// others are all normal tenants.
	IsAdmin bool
}

// Business represents a cmdb business under a tenant.
type Business struct {
	// belongs to.
	TenantID string

	// biz-id is the unique identifier for a business.
	BizID   int64
	BizName string
}

// NetArea represents a cmdb net-area. In which IPs will not be duplicated.
type NetArea struct {
	// belongs to
	TenantID string

	// cloud-id is the unique identifier for a net-area.
	CloudID int64
}

// Host represents a cmdb host.
type Host struct {
	// belongs to
	TenantID string
	CloudID  int64
	BizID    int64

	// host-id is the unique identifier for a host.
	HostID int64

	// host information.
	InnerIP string
	Mac     string
	OSType  string
}
