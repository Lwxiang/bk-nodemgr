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
	"time"

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/manager/workflowdef"
	types "git.woa.com/bk-gse/bk-nodeman/internal/backend/types/router/api-v3"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
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

	err := h.manager.ExecuteOperation(workflowdef.OperDefNameSyncBizAndHost, "trigger-1", &operengine.OperInstParam{
		Timeout:     1 * time.Minute,
		InitContent: map[string]map[string]any{},
	})
	if err != nil {
		h.logger.Errorf("failed to start sync cmdb host operation, err: %v", err)
		return nil, err
	}

	h.logger.Infof("successfully started sync cmdb host operation.")

	resp := new(types.SyncCmdbHostResp)

	return resp, nil
}
