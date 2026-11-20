/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package topo

import (
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ListCloudSecurityGroupPools lists globally configured pools visible to managers.
func (h *handler) ListCloudSecurityGroupPools(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.TopoCloudSecurityGroupPoolListReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list cloud security group pools, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := h.authorizer.Check(rCtx, auth.ActionCloudSecurityGroupManage, nil); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to list cloud security group pools, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	resp := new(protoBackend.TopoCloudSecurityGroupPoolListResp)
	resp.ConvertPoolsFromTypes(h.cloudSGRegistry.ListPools())

	return resp.GetData(), nil
}

func (h *handler) validateAccessPointPools(accessPoints []*types.AccessPoint) error {
	for _, accessPoint := range accessPoints {
		poolID := accessPoint.CloudSecurityGroupPoolID
		if poolID == "" {
			continue
		}
		if !h.cloudSGRegistry.Has(poolID) {
			return fmt.Errorf("accesspoint %q: unknown cloud security group pool: %s", accessPoint.Name, poolID)
		}
	}

	return nil
}

func hasPoolBinding(accessPoints []*types.AccessPoint) bool {
	for _, accessPoint := range accessPoints {
		if accessPoint.CloudSecurityGroupPoolID != "" {
			return true
		}
	}

	return false
}

func hasPoolBindingChanged(accessPoints []*types.AccessPoint, existing map[int64]string) bool {
	for _, accessPoint := range accessPoints {
		if accessPoint.CloudSecurityGroupPoolID != existing[accessPoint.ID] {
			return true
		}
	}

	return false
}

func (h *handler) checkCloudSGManage(rCtx restserver.IContext) error {
	if err := h.authorizer.Check(rCtx, auth.ActionCloudSecurityGroupManage, nil); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to change cloud security group binding, permission denied")
		return resterrf.ErrWrap(resterrf.PermissionDenied, err)
	}

	return nil
}

func (h *handler) listExistingCloudSGPoolIDs(
	nCtx contextx.IContext, accessPointIDs []int64,
) (map[int64]string, error) {

	result := make(map[int64]string)
	if len(accessPointIDs) == 0 {
		return result, nil
	}

	accessPoints, _, err := h.storage.ListAccessPoint(nCtx, types.UnlimitedPage(), &types.AccessPointCondition{
		ExactInclude: &types.AccessPointExactFields{AccessPointID: accessPointIDs},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list existing accesspoint cloud security group pools: %w", err)
	}

	for _, accessPoint := range accessPoints {
		result[accessPoint.ID] = accessPoint.CloudSecurityGroupPoolID
	}

	return result, nil
}
