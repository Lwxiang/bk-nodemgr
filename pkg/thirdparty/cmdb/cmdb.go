/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package cmdb provides handlers to operate cmd api.
package cmdb

import (
	"context"
	"fmt"
	"net/http"

	"git.woa.com/bk-gse/bk-nodeman/pkg/rest"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/client"
	restheader "git.woa.com/bk-gse/bk-nodeman/pkg/rest/header"
)

// This file only supports requesting and getting responses.

type HeaderSetter interface {
	GetAuthHeader() (string, error)
}

// Config the config of cmdb.
type Config struct {
	TenantID     string
	HeaderSetter HeaderSetter
}

// cli client for cmdb.
type cli struct {
	client rest.ClientInterface
	config *Config
}

// newClient initialize a new cmdb client.
func newClient(c *client.Capability, conf *Config) (*cli, error) {
	restCli, err := rest.NewClient(c, "/api/v3")
	if err != nil {
		return nil, err
	}

	return &cli{
		client: restCli,
		config: conf,
	}, nil
}

// getCommonHeader get cmdb common header.
func (c *cli) getCommonHeader() (http.Header, error) {
	header := http.Header{}
	header.Set(restheader.RIDKey, restheader.RIDGenerator())

	// TODO: 接入租户信息

	authHeader, err := c.config.HeaderSetter.GetAuthHeader()
	if err != nil {
		return nil, err
	}

	header.Set(restheader.BKGWAuthKey, authHeader)

	return header, nil
}

// ListBizHosts ...
func (c *cli) listBizHosts(ctx context.Context, req *ListBizHostsReq) (*ListBizHostsResp, error) {
	resp := new(BaseBroker[*ListBizHostsResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/hosts/app/%d/list_hosts", req.BKBizID).
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("list biz hosts failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchBusiness search cmdb business.
func (c *cli) searchBusiness(ctx context.Context, req *SearchBusinessReq) (*SearchBusinessResp, error) {
	resp := new(BaseBroker[*SearchBusinessResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	// TODO: access tenant information.
	err = c.client.Post().
		SubResourcef("/biz/search/%s", "0").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search business failed, err: %v", err)
	}

	return resp.Data, nil
}

// searchCloudArea search cloud area.
func (c *cli) searchCloudArea(ctx context.Context, req *SearchCloudAreaReq) (*SearchCloudAreaResp, error) {
	resp := new(BaseBroker[*SearchCloudAreaResp])
	header, err := c.getCommonHeader()
	if err != nil {
		return nil, err
	}

	err = c.client.Post().
		SubResourcef("/findmany/cloudarea").
		WithContext(ctx).
		WithHeaders(header).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if err := resp.IsFailed(); err != nil {
		return nil, fmt.Errorf("search cloud area failed, err: %v", err)
	}

	return resp.Data, nil
}
