/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"time"

	baseStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Tenant represents a tenant.
// TenantID should be the unique key.
type Tenant struct {
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	IsAdmin  bool   `json:"is_admin"`
}

// TableTenant represents the complete db structures of a tenant.
type TableTenant struct {
	baseStorage.BasicInfo `json:"basic" bson:"basic"`
	Data                  Tenant `json:"data" bson:"data"`
}

// Business represents a business under a tenant.
// BizID should be the unique key.
type Business struct {
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	BizID    int    `json:"biz_id" bson:"biz_id"`
	BizName  string `json:"biz_name" bson:"biz_name"`
}

func (b *Business) fromRuntime(biz *types.Business) {
	b.TenantID = biz.TenantID
	b.BizID = biz.BizID
	b.BizName = biz.BizName
}

func (b *Business) toRuntime() *types.Business {
	return &types.Business{
		TenantID: b.TenantID,
		BizID:    b.BizID,
		BizName:  b.BizName,
	}
}

// TableBusiness represents the complete db structures of a business.
type TableBusiness struct {
	baseStorage.BasicInfo `json:"basic" bson:"basic"`
	Data                  *Business `json:"data" bson:"data"`
}

func (tb *TableBusiness) indexes() []*mongo.IndexModel {
	return []*mongo.IndexModel{

		// biz_id should be unique.
		{
			Keys:    bson.D{{Key: "data.biz_id", Value: 1}},
			Options: mongoOptions.Index().SetUnique(true),
		},
	}
}

// upsertParams returns the filter, update and options for upserting a business.
func (tb *TableBusiness) upsertParams(biz *Business) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
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

// NetArea represents a network area, in which IPs will not be duplicated.
type NetArea struct {
	Name    string `json:"name" bson:"name"`
	CloudID int    `json:"cloud_id" bson:"cloud_id"`
}

// TableNetArea represents the complete db structures of a net area.
type TableNetArea struct {
	baseStorage.BasicInfo `json:"basic" bson:"basic"`
	Data                  NetArea `json:"data" bson:"data"`
}

// Host represents a host.
type Host struct {
	TenentID string `json:"tenant_id" bson:"tenant_id"`
	CloudID  int    `json:"cloud_id" bson:"cloud_id"`
	BizID    string `json:"biz_id" bson:"biz_id"`
	HostID   int    `json:"host_id" bson:"host_id"`
	InnerIP  string `json:"inner_ip" bson:"inner_ip"`
	Mac      string `json:"mac" bson:"mac"`
	OSType   string `json:"os_type" bson:"os_type"`
}

// TableHost represents the complete db structures of a host.
type TableHost struct {
	baseStorage.BasicInfo `json:"basic" bson:"basic"`
	Data                  Host `json:"data" bson:"data"`
}
