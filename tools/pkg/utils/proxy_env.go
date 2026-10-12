/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package utils

import (
	"fmt"
	"os"
)

// proxyEnvKeys lists the environment variables consulted by http.ProxyFromEnvironment.
var proxyEnvKeys = []string{
	"http_proxy", "https_proxy",
	"HTTP_PROXY", "HTTPS_PROXY",
}

// OverrideProxyEnv overrides the proxy environment variables, so that all http requests
// made by this process use the given proxy. An empty proxy value means explicitly skipping
// the proxy. It must be called before any http request, because http.ProxyFromEnvironment
// caches the environment on its first evaluation.
func OverrideProxyEnv(proxy string) error {
	if proxy == "" {
		return UnsetProxyEnv()
	}

	for _, key := range proxyEnvKeys {
		if err := os.Setenv(key, proxy); err != nil {
			return fmt.Errorf("failed to set env %s: %w", key, err)
		}
	}

	return nil
}

// UnsetProxyEnv unsets the proxy environment variables, so that all http requests
// made by this process skip the proxy. It must be called before any http request.
func UnsetProxyEnv() error {
	for _, key := range proxyEnvKeys {
		if err := os.Unsetenv(key); err != nil {
			return fmt.Errorf("failed to unset env %s: %w", key, err)
		}
	}

	return nil
}
