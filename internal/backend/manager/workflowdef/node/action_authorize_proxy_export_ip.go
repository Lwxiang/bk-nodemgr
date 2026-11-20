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

package node

import (
	"fmt"
	"time"

	cloudsgSvc "github.com/TencentBlueKing/bk-nodemgr/internal/backend/cloudsg"
	nodeUtils "github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/workflowdef/node/utils"
	nodeStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/node"
	topoStg "github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/topo"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

const (
	// ActionNameAuthorizeProxyExportIP defines the action name.
	ActionNameAuthorizeProxyExportIP = "authorize_proxy_export_ip"
)

// NewActionAuthorizeProxyExportIP creates a new action.
func NewActionAuthorizeProxyExportIP(capability *Capability) action.Definition {
	return &actionAuthorizeProxyExportIP{
		storageTopo:     capability.StorageTopo,
		storageNode:     capability.StorageNode,
		cloudSGRegistry: capability.CloudSGRegistry,
	}
}

// ActParamAuthorizeProxyExportIP defines the action parameter.
type ActParamAuthorizeProxyExportIP struct {
	nodeUtils.NodeActionStandardParam `json:",inline"`
}

type actionAuthorizeProxyExportIP struct {
	storageTopo     topoStg.IStorage
	storageNode     nodeStg.IDaoNodeDeployment
	cloudSGRegistry cloudsgSvc.IRegistry
}

// Name returns the name of the action.
func (act *actionAuthorizeProxyExportIP) Name() string {
	return ActionNameAuthorizeProxyExportIP
}

// DisplayNameZh returns the Chinese display name of the action.
func (act *actionAuthorizeProxyExportIP) DisplayNameZh() string {
	return "放行Proxy出口IP至云安全组"
}

// DisplayNameEn returns the English display name of the action.
func (act *actionAuthorizeProxyExportIP) DisplayNameEn() string {
	return "Authorize Proxy Export IP to Cloud Security Groups"
}

// Version returns the version of the action.
func (act *actionAuthorizeProxyExportIP) Version() string {
	return "v1.0.0" // nolint: goconst
}

// Description returns the description of the action.
func (act *actionAuthorizeProxyExportIP) Description() string {
	return "authorize the proxy export ip to the cloud security groups bound to the upstream access points"
}

// Timeout returns the timeout of the action.
func (act *actionAuthorizeProxyExportIP) Timeout() time.Duration {
	return 1 * time.Minute
}

// Tags returns the tags of the action.
func (act *actionAuthorizeProxyExportIP) Tags() []action.Tag {
	return []action.Tag{}
}

// Queue returns the queue of the action.
func (act *actionAuthorizeProxyExportIP) Queue() action.Queue {
	return action.QueueDefault
}

// MaxRetryCount returns the max retry count of the action.
func (act *actionAuthorizeProxyExportIP) MaxRetryCount() uint {
	return 3 //nolint: mnd
}

// DelayFn defines the delay before retrying on failure.
func (act *actionAuthorizeProxyExportIP) DelayFn(attempt int) func() {
	return action.DefaultBackoffDelayFn(act, attempt)
}

// Do executes the action.
//
// The action only adds rules, it never deletes. Failures abort the install on purpose:
// the proxy cannot reach the upstream without the authorization, so this action sits on
// the critical path. The error message carries the pool and security group ids so that
// operators can finish or fix the authorization manually in the cloud console.
func (act *actionAuthorizeProxyExportIP) Do(ctx *action.InstanceContext) error {
	param := new(ActParamAuthorizeProxyExportIP)
	if err := conv.MapToStruct(ctx.Data.Content, param); err != nil {
		return err
	}

	std := nodeUtils.NewNodeActionStandarder(act.storageNode, nil)
	if err := std.Initialize(ctx, param.NodeActionStandardParam); err != nil {
		return err
	}

	nCtx := std.Context()
	deployInfo := std.DeployInfo()

	// only proxies need the export ip authorization, agent and other roles skip this action.
	if deployInfo.Host.Dynamic.NodeRole != types.NodeRoleProxy {
		return nil
	}

	bindings, err := act.resolveUpstreamBindings(nCtx, deployInfo.Host.Dynamic.NetworkUnitID)
	if err != nil {
		return err
	}

	if len(bindings) == 0 {
		return nil
	}

	exportIP := deployInfo.Host.Dynamic.ExportIP
	if exportIP == "" {
		return fmt.Errorf("proxy export ip is empty, cannot authorize cloud security groups, host-id(%d)",
			deployInfo.Host.HostID)
	}

	for _, binding := range bindings {
		if err := act.authorize(nCtx, binding, exportIP); err != nil {
			return err
		}

		// log the exact target and rule of this authorization, normalized the same way the
		// registry writes it, for example ALL ports when the pool enables portAll.
		rule := cloudsg.CloudSecurityGroupRule{
			CIDR:     exportIP + "/32",
			Port:     cloudsgSvc.CombinePorts(binding.ports),
			Protocol: "TCP",
			Remark:   fmt.Sprintf("na=%d|nu=%d", binding.networkAreaID, binding.networkUnitID),
		}
		rule = act.cloudSGRegistry.NormalizeRules(binding.poolID, []cloudsg.CloudSecurityGroupRule{rule})[0]
		securityGroupIDs := act.cloudSGRegistry.SecurityGroupIDs(binding.poolID)

		std.InstanceData().Log().
			Zh("放行Proxy出口IP至云安全组成功, 主机ID(%d), 出口IP(%s), 安全组池(%s), 安全组(%v), "+
				"规则(来源CIDR=%s, 端口=%s, 协议=%s, 备注=%s)",
				deployInfo.Host.HostID, exportIP, binding.poolID, securityGroupIDs,
				rule.CIDR, rule.Port, rule.Protocol, rule.Remark).
			En("successfully authorized proxy export ip, host-id(%d), export-ip(%s), pool(%s), security-groups(%v), "+
				"rule(cidr=%s, ports=%s, protocol=%s, remark=%s)",
				deployInfo.Host.HostID, exportIP, binding.poolID, securityGroupIDs,
				rule.CIDR, rule.Port, rule.Protocol, rule.Remark).
			Info()
	}

	return nil
}

// upstreamBinding pairs one enabled cloud security group binding with the endpoint ports
// of its access point.
type upstreamBinding struct {
	accessPointID int64
	networkAreaID int64
	networkUnitID int64
	poolID        string
	ports         []string
}

// resolveUpstreamBindings loads the enabled cloud security group bindings of the access
// points referenced by the network unit upstream links. Units without enabled bindings
// make the action a no-op.
func (act *actionAuthorizeProxyExportIP) resolveUpstreamBindings(
	nCtx contextx.IContext, networkUnitID int64,
) ([]*upstreamBinding, error) {

	if networkUnitID <= 0 {
		return nil, nil
	}

	networkUnit, err := act.storageTopo.GetNetworkUnit(nCtx, networkUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get networkunit(%d): %w", networkUnitID, err)
	}

	if networkUnit == nil || networkUnit.IsDirect {
		return nil, nil
	}

	accessPointIDs := uniqueAccessPointIDs(networkUnit.Links)
	if len(accessPointIDs) == 0 {
		return nil, nil
	}

	accessPoints, _, err := act.storageTopo.ListAccessPoint(nCtx, types.UnlimitedPage(), &types.AccessPointCondition{
		ExactInclude: &types.AccessPointExactFields{AccessPointID: accessPointIDs},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list accesspoints of networkunit(%d): %w", networkUnitID, err)
	}

	bindings := make([]*upstreamBinding, 0, len(accessPoints))
	for _, accessPoint := range accessPoints {
		poolID := accessPoint.CloudSecurityGroupPoolID
		if poolID == "" {
			continue
		}

		ports, err := cloudsgSvc.EndpointPorts(accessPoint.Endpoints)
		if err != nil {
			return nil, fmt.Errorf("accesspoint(%d) %q: %w", accessPoint.ID, accessPoint.Name, err)
		}

		bindings = append(bindings, &upstreamBinding{
			accessPointID: accessPoint.ID,
			networkAreaID: accessPoint.NetworkAreaID,
			networkUnitID: networkUnitID,
			poolID:        poolID,
			ports:         ports,
		})
	}

	return bindings, nil
}

// authorize adds the export ip rule to one bound global pool.
func (act *actionAuthorizeProxyExportIP) authorize(
	nCtx contextx.IContext, upstream *upstreamBinding, exportIP string,
) error {

	// one combined rule per export ip: the discrete port expression keeps the rule quota
	// usage at one entry per ip instead of one entry per endpoint port.
	rules := []cloudsg.CloudSecurityGroupRule{
		{
			CIDR:     exportIP + "/32",
			Port:     cloudsgSvc.CombinePorts(upstream.ports),
			Protocol: "TCP",
			Remark:   fmt.Sprintf("na=%d|nu=%d", upstream.networkAreaID, upstream.networkUnitID),
		},
	}

	if err := act.cloudSGRegistry.Authorize(nCtx, upstream.poolID, rules); err != nil {
		return fmt.Errorf("failed to authorize export ip %s/32, accesspoint(%d), pool(%s), "+
			"finish or fix the authorization manually in the cloud console: %w",
			exportIP, upstream.accessPointID, upstream.poolID, err)
	}

	return nil
}

// uniqueAccessPointIDs collects the unique access point ids referenced by the links.
func uniqueAccessPointIDs(links types.Links) []int64 {
	seen := make(map[int64]struct{})
	ids := make([]int64, 0, 3) //nolint: mnd

	for _, link := range []*types.Link{links.Cluster, links.File, links.Data} {
		if link == nil {
			continue
		}

		if _, ok := seen[link.AccessPointID]; ok {
			continue
		}

		seen[link.AccessPointID] = struct{}{}
		ids = append(ids, link.AccessPointID)
	}

	return ids
}
