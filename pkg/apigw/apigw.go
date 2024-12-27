/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package apigw provides an APIGateway client.
package apigw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.woa.com/bk-gse/bk-nodeman/pkg/auth"
)

type Client interface {
	Get(userAuth *auth.UserAuth, url string, pathParameters map[string]string) (int, []byte, error)
	Post(userAuth *auth.UserAuth, url string, body []byte) (int, []byte, error)
}

type Config struct {
	BKAppCode        string `json:"bk_app_code"`
	BKAppSecret      string `json:"bk_app_secret"`
	PlatformUsername string `json:"bk_username"`

	APIGWDomain string
	SSMDomain   string
}

func NewClient(config *Config) Client {
	return &client{
		config:  config,
		httpCli: &http.Client{},
	}
}

// client provide an APIGateway client.
type client struct {
	config *Config

	httpCli *http.Client
}

func (c *client) Get(userAuth *auth.UserAuth, url string, pathParameters map[string]string) (int, []byte, error) {
	return 0, nil, nil
}

func (c *client) Post(userAuth *auth.UserAuth, url string, body []byte) (int, []byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", c.config.APIGWDomain, url), bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	c.generateAuth(userAuth, req)
	return c.do(req)
}

func (c *client) Put(url string, body string) (*http.Response, error) {
	return nil, nil
}

func (c *client) Delete(url string, body string) (*http.Response, error) {
	return nil, nil
}

func (c *client) do(req *http.Request) (int, []byte, error) {
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to do request. code(%d), url(%s), err: %v", resp.StatusCode, req.URL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return resp.StatusCode, nil, fmt.Errorf("got HTTP response code(%d), url(%s)", resp.StatusCode, req.URL.String())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response body. code(%d), url(%s), err: %v", resp.StatusCode, req.URL.String(), err)
	}

	fmt.Printf("response body: %s\n", data)

	return resp.StatusCode, data, nil
}

func (c *client) generateAuth(userAuth *auth.UserAuth, req *http.Request) error {
	if userAuth == nil {
		authHeader, _ := json.Marshal(&AuthHeader{
			BKAppCode:   c.config.BKAppCode,
			BKAppSecret: c.config.BKAppSecret,
			BKUsername:  c.config.PlatformUsername,
		})
		req.Header.Set("X-Bkapi-Authorization", string(authHeader))

		return nil
	}

	return fmt.Errorf("not implemented")
}
