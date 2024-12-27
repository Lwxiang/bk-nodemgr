/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides handlers to operate cmd API.
package cmdb

import (
	"encoding/json"
	"fmt"

	"git.woa.com/bk-gse/bk-nodeman/pkg/apigw"
	"git.woa.com/bk-gse/bk-nodeman/pkg/auth"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
)

// Handler is the interface for cmdb handler.
type Handler interface {
	Tenant(tenantID int) TenantClient
}

// TenantClient is the interface for cmdb tenant client.
type TenantClient interface {
	User(userAuth *auth.UserAuth) Client
}

// Client is the interface for cmdb client.
type Client interface {
	// ListBizHost lists the hosts of a specified business.
	ListBizHost(page Page, bizID int) (*RespListBizHosts, error)

	// SearchBusiness get all the business in cmdb.
	SearchBusiness(page Page) ([]*types.Business, int, error)
}

// Config defines the config for cmdb handler.
type Config struct {
	Environment string
}

// NewHandler creates a new cmdb handler.
func NewHandler(config *Config, apigwCli apigw.Client) Handler {
	return &handler{
		config:   config,
		apigwCli: apigwCli,
	}
}

type handler struct {
	config   *Config
	apigwCli apigw.Client
}

func (h *handler) Tenant(tenantID int) TenantClient {
	return &tenantClient{
		hld:      h,
		tenantID: tenantID,
	}
}

type tenantClient struct {
	hld *handler

	tenantID int
}

func (t *tenantClient) User(userAuth *auth.UserAuth) Client {
	return &client{
		hld:    t.hld,
		tenant: t,
		auth:   userAuth,
	}
}

type client struct {
	hld    *handler
	tenant *tenantClient

	auth *auth.UserAuth
}

func (c *client) ListBizHost(page Page, bizID int) (*RespListBizHosts, error) {
	body, _ := json.Marshal(&ReqListBizHosts{
		Page:    page,
		BKBizID: bizID,
		Fields: []string{
			"bk_host_innerip",
			"bk_cloud_id",
			"bk_host_id",
			"bk_module_id",
		},
	})

	url := fmt.Sprintf("api/bk-cmdb/%s/api/v3/hosts/app/%d/list_hosts", c.hld.config.Environment, bizID)
	code, data, err := c.hld.apigwCli.Post(c.auth, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to do post(%s). http-code(%d), err: %v", url, code, err)
	}

	resp := &RespListBizHosts{}
	if err = json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, data, err)
	}

	return resp, nil
}

func (c *client) SearchBusiness(page Page) ([]*types.Business, int, error) {
	body, _ := json.Marshal(&ReqSearchBusiness{
		Page: page,
	})

	url := fmt.Sprintf("api/bk-cmdb/%s/api/v3/biz/search/%s", c.hld.config.Environment, "blueking")
	code, respData, err := c.hld.apigwCli.Post(c.auth, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to do post(%s). http-code(%d), err: %v", url, code, err)
	}

	resp := &RespSearchBusiness{}
	if err = json.Unmarshal(respData, &resp); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, respData, err)
	}

	data := make([]*types.Business, 0, len(resp.Data.Info))
	for _, info := range resp.Data.Info {
		data = append(data, &types.Business{
			TenantID: "", // TODO: get tenant-id from cmdb
			BizID:    info.BKBizID,
			BizName:  info.BKBizName,
		})
	}

	return data, resp.Data.Count, nil
}
