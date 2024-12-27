/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package apigw ...
package apigw

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Identity    struct {
		UserType string `json:"user_type"`
		Username string `json:"username"`
	} `json:"identity"`
	RefreshToken string `json:"refresh_token"`
}

type AuthHeader struct {
	BKAppCode   string `json:"bk_app_code"`
	BKAppSecret string `json:"bk_app_secret"`
	BKUsername  string `json:"bk_username"`
}

type ReqSSMAccessToken struct {
	GrantType  string `json:"grant_type"`
	IDProvider string `json:"id_provider"`
	BKToken    string `json:"bk_token"`
}

type RespSSMAccessToken struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *AccessToken `json:"data"`
}
