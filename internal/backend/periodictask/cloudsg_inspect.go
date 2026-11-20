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

package periodictask

import (
	"fmt"
	"time"

	cloudsgSvc "github.com/TencentBlueKing/bk-nodemgr/internal/backend/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	periodicTaskTimeoutCloudSGInspect = 10 * time.Minute
	periodicTaskNameCloudSGInspect    = "cloudsg_inspect"
)

// CloudSGInspect inspects global pools against desired state aggregated from every tenant's topology.
// It is strictly read-only: drifts and quota usage are only logged.
func (pt *PeriodicTask) CloudSGInspect(nCtx contextx.IContext) error {
	pools := pt.conf.CloudSG.ListPools()
	if len(pools) == 0 {
		return nil
	}

	desiredByPool := make(map[string][]cloudsg.CloudSecurityGroupRule)
	tenantIDs, err := tenant.ListEnabledTenantIDs(nCtx)
	if err != nil {
		return fmt.Errorf("failed to list enabled tenants: %w", err)
	}

	for _, tenantID := range tenantIDs {
		tenantCtx := contextx.From(nCtx, contextx.WithTenantID(tenantID))
		if err := pt.collectDesiredRules(tenantCtx, desiredByPool); err != nil {
			return fmt.Errorf("failed to derive cloud security group desired state for tenant(%s): %w", tenantID, err)
		}
	}

	pt.inspectPools(nCtx, pools, desiredByPool)

	return nil
}

func (pt *PeriodicTask) collectDesiredRules(
	nCtx contextx.IContext, desiredByPool map[string][]cloudsg.CloudSecurityGroupRule,
) error {

	accessPoints, err := pt.listEnabledCloudSGAccessPoints(nCtx)
	if err != nil {
		return err
	}
	if len(accessPoints) == 0 {
		return nil
	}

	upstreamUnits, err := pt.mapUpstreamUnitsByAccessPoint(nCtx)
	if err != nil {
		return err
	}

	for _, accessPoint := range accessPoints {
		poolID := accessPoint.CloudSecurityGroupPoolID
		if !pt.conf.CloudSG.Has(poolID) {
			return fmt.Errorf("accesspoint(%d) %q references unknown cloud security group pool: %s",
				accessPoint.ID, accessPoint.Name, poolID)
		}

		desired, err := pt.desiredRules(nCtx, accessPoint, upstreamUnits[accessPoint.ID])
		if err != nil {
			return fmt.Errorf("failed to derive desired rules for accesspoint(%d): %w", accessPoint.ID, err)
		}

		// apply the pool level transformations so the desired state matches what the
		// authorize action actually writes, for example ALL ports when the pool enables portAll.
		desired = pt.conf.CloudSG.NormalizeRules(poolID, desired)
		desiredByPool[poolID] = append(desiredByPool[poolID], desired...)
	}

	return nil
}

func (pt *PeriodicTask) listEnabledCloudSGAccessPoints(nCtx contextx.IContext) ([]*types.AccessPoint, error) {
	accessPoints, _, err := pt.conf.StgTopo.ListAccessPoint(nCtx, types.UnlimitedPage())
	if err != nil {
		return nil, fmt.Errorf("failed to list accesspoints: %w", err)
	}

	enabled := make([]*types.AccessPoint, 0, len(accessPoints))
	for _, accessPoint := range accessPoints {
		if accessPoint.CloudSecurityGroupPoolID != "" {
			enabled = append(enabled, accessPoint)
		}
	}

	return enabled, nil
}

func (pt *PeriodicTask) mapUpstreamUnitsByAccessPoint(nCtx contextx.IContext) (map[int64][]int64, error) {
	networkUnits, _, err := pt.conf.StgTopo.ListNetworkUnit(nCtx, types.UnlimitedPage())
	if err != nil {
		return nil, fmt.Errorf("failed to list networkunits: %w", err)
	}

	result := make(map[int64][]int64)
	for _, unit := range networkUnits {
		if unit == nil {
			continue
		}

		for _, link := range []*types.Link{unit.Links.Cluster, unit.Links.File, unit.Links.Data} {
			if link == nil || link.AccessPointID < 0 {
				continue
			}

			if containsInt64(result[link.AccessPointID], unit.ID) {
				continue
			}

			result[link.AccessPointID] = append(result[link.AccessPointID], unit.ID)
		}
	}

	return result, nil
}

func (pt *PeriodicTask) inspectPools(
	nCtx contextx.IContext,
	pools []cloudsg.CloudSecurityGroupPoolSummary,
	desiredByPool map[string][]cloudsg.CloudSecurityGroupRule,
) {

	for _, pool := range pools {
		inspections, err := pt.conf.CloudSG.Inspect(nCtx, pool.PoolID)
		if err != nil {
			logger.G.Sys().WithErr(err).With("pool-id", pool.PoolID).
				Warn("failed to inspect cloud security group pool")

			continue
		}

		desired := desiredByPool[pool.PoolID]
		log := logger.G.Sys().With(
			"pool-id", pool.PoolID,
			"pool-name", pool.PoolName,
			"desired-rule-count", len(desired),
		)

		for _, inspection := range inspections {
			missing, stale := cloudsgSvc.DiffRules(desired, inspection.Rules)
			log.With(
				"security-group-id", inspection.SecurityGroupID,
				"total-ingress", inspection.TotalIngress,
				"rule-count", len(inspection.Rules),
				"missing-rules", fmt.Sprintf("%+v", missing),
				"stale-rules", fmt.Sprintf("%+v", stale),
			).Info("inspected cloud security group pool")
		}
	}
}

func (pt *PeriodicTask) desiredRules(
	nCtx contextx.IContext, accessPoint *types.AccessPoint, upstreamUnitIDs []int64,
) ([]cloudsg.CloudSecurityGroupRule, error) {

	if len(upstreamUnitIDs) == 0 {
		return nil, nil
	}

	hosts, _, err := pt.conf.StgTopo.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
		DynamicExactInclude: &types.HostDynamicExactFields{
			NetworkUnitID: upstreamUnitIDs,
			NodeRole:      []types.NodeRole{types.NodeRoleProxy},
			NodeStatus:    []types.NodeStatus{types.NodeStatusRunning},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list proxies: %w", err)
	}

	ports, err := cloudsgSvc.EndpointPorts(accessPoint.Endpoints)
	if err != nil {
		return nil, fmt.Errorf("accesspoint(%d) %q: %w", accessPoint.ID, accessPoint.Name, err)
	}

	// desired rules use the same combined port expression as the authorize action,
	// keeping the diff consistent with what the action actually writes.
	combinedPorts := cloudsgSvc.CombinePorts(ports)

	seen := make(map[string]struct{})
	rules := make([]cloudsg.CloudSecurityGroupRule, 0, len(hosts))
	for _, host := range hosts {
		exportIP := host.Dynamic.ExportIP
		if exportIP == "" {
			continue
		}

		if _, ok := seen[exportIP]; ok {
			continue
		}

		seen[exportIP] = struct{}{}
		rules = append(rules, cloudsg.CloudSecurityGroupRule{
			CIDR:     exportIP + "/32",
			Port:     combinedPorts,
			Protocol: "TCP",
		})
	}

	return rules, nil
}

func containsInt64(values []int64, value int64) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}

	return false
}
