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

package cloudsg

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// EndpointPorts returns unique ports from access point endpoints.
func EndpointPorts(endpoints types.Endpoints) ([]string, error) {
	seen := make(map[string]struct{})
	ports := make([]string, 0, 3) //nolint:mnd

	for _, group := range [][]string{endpoints.Cluster, endpoints.File, endpoints.Data} {
		for _, endpoint := range group {
			_, port, err := net.SplitHostPort(endpoint)
			if err != nil {
				return nil, fmt.Errorf("invalid endpoint %q: %w", endpoint, err)
			}

			if _, ok := seen[port]; ok {
				continue
			}

			seen[port] = struct{}{}
			ports = append(ports, port)
		}
	}

	return ports, nil
}

// CombinePorts merges ports into one discrete port expression supported by the cloud
// security group api, for example "28625,28668,28925". One combined rule per export ip
// costs one third of the rule quota of per-port rules. Ports are deduplicated and sorted
// so the expression is stable for idempotent comparison.
func CombinePorts(ports []string) string {
	seen := make(map[string]struct{}, len(ports))
	unique := make([]string, 0, len(ports))
	for _, port := range ports {
		if _, ok := seen[port]; ok {
			continue
		}

		seen[port] = struct{}{}
		unique = append(unique, port)
	}

	sort.Strings(unique)

	return strings.Join(unique, ",")
}

// DiffRules splits desired and actual rules into missing and stale rules.
func DiffRules(desired, actual []cloudsg.CloudSecurityGroupRule) (
	[]cloudsg.CloudSecurityGroupRule, []cloudsg.CloudSecurityGroupRule,
) {

	missing := make([]cloudsg.CloudSecurityGroupRule, 0)
	for _, rule := range desired {
		// a desired rule already covered by any live entry is not missing, matching the
		// coverage-aware skip of the authorize path.
		if cloudsg.CloudSecurityGroupRuleCoveredBy(actual, rule) {
			continue
		}

		missing = append(missing, rule)
	}

	desiredSet := make(map[string]cloudsg.CloudSecurityGroupRule, len(desired))
	for _, rule := range desired {
		desiredSet[ruleKey(rule)] = rule
	}

	stale := make([]cloudsg.CloudSecurityGroupRule, 0)
	for _, rule := range actual {
		if _, ok := desiredSet[ruleKey(rule)]; !ok {
			stale = append(stale, rule)
		}
	}

	return missing, stale
}

func ruleKey(rule cloudsg.CloudSecurityGroupRule) string {
	return rule.CIDR + "|" + rule.Port + "|" + strings.ToUpper(rule.Protocol) + "|" + string(rule.EffectiveAction())
}
