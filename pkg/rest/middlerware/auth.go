/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package middleware Authorization
package middleware

import (
	"net/http"

	"git.woa.com/bk-gse/bk-nodeman/pkg/rest"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/errf"
	"github.com/gin-gonic/gin"
)

// AuthenticationRequired Authorization
func AuthenticationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext := rest.InitRestContext(c)

		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		switch {
		case initContextWithJWT(restContext):
		default:
			rest.AbortWithUnauthorizedError(restContext, errf.ErrorUnauthorized)
			return
		}

		c.Next()
	}
}

// initContextWithJWT init context with jwt
func initContextWithJWT(c *rest.Context) bool {
	// TODO: implement jwt
	return true
}
