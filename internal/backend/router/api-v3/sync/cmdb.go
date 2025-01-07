/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package sync ...
package sync

import (
	"git.woa.com/bk-gse/bk-nodeman/internal/backend/criteria/constant"
	"git.woa.com/bk-gse/bk-nodeman/internal/backend/manager"
	types "git.woa.com/bk-gse/bk-nodeman/internal/backend/types/router/api-v3"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest"
)

// SyncCmdbHost ...
func (h *handler) SyncCmdbHost(ctx *rest.Context) (interface{}, error) {
	req := new(types.SyncCmdbHostReq)
	if err := ctx.BindJSON(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := h.manager.StartPipeline(manager.PipelineSyncFromCmdb, constant.PipelineTimeoutDefault)
	if err != nil {
		blog.Errorf("failed to start sync cmdb host pipeline, err: %v", err)
		return nil, err
	}

	blog.Infof("successfully started sync cmdb host pipeline.")

	resp := new(types.SyncCmdbHostResp)

	return resp, nil
}
