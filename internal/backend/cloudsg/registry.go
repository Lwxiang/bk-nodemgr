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

// Package cloudsg manages globally configured cloud security group pools.
package cloudsg

import (
	"errors"
	"fmt"
	"sort"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/cloudsg"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/locker"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/tencentcloud"
)

// IRegistry exposes globally configured cloud security group pools.
type IRegistry interface {
	ListPools() []cloudsg.CloudSecurityGroupPoolSummary
	Has(poolID string) bool
	Authorize(nCtx contextx.IContext, poolID string, rules []cloudsg.CloudSecurityGroupRule) error
	Inspect(nCtx contextx.IContext, poolID string) ([]cloudsg.CloudSecurityGroupInspection, error)
	// NormalizeRules applies the pool level rule transformations, keeping the authorize
	// path and the inspection desired state consistent.
	NormalizeRules(poolID string, rules []cloudsg.CloudSecurityGroupRule) []cloudsg.CloudSecurityGroupRule
	// SecurityGroupIDs returns the security group ids of one pool for result logging.
	SecurityGroupIDs(poolID string) []string
}

// Registry is an immutable in-memory registry of global pools.
type Registry struct {
	pools     map[string]config.CloudSecurityGroupPool
	summaries []cloudsg.CloudSecurityGroupPoolSummary
	handlers  map[cloudsg.CloudSecurityGroupVendor]tencentcloud.ISecurityGroupHandler
	locker    locker.MutexFactory
}

// NewRegistry builds the global pool registry.
func NewRegistry(
	conf config.CloudSecurityGroup,
	handlers map[cloudsg.CloudSecurityGroupVendor]tencentcloud.ISecurityGroupHandler,
	lockerFactory locker.MutexFactory,
) IRegistry {

	registry := &Registry{
		pools:     make(map[string]config.CloudSecurityGroupPool, len(conf.Pools)),
		summaries: make([]cloudsg.CloudSecurityGroupPoolSummary, 0, len(conf.Pools)),
		handlers:  handlers,
		locker:    lockerFactory,
	}

	for _, pool := range conf.Pools {
		pool.SecurityGroupIDs = append([]string(nil), pool.SecurityGroupIDs...)
		registry.pools[pool.PoolID] = pool
		registry.summaries = append(registry.summaries, cloudsg.CloudSecurityGroupPoolSummary{
			PoolID:   pool.PoolID,
			PoolName: pool.PoolName,
		})
	}

	sort.Slice(registry.summaries, func(i, j int) bool {
		return registry.summaries[i].PoolID < registry.summaries[j].PoolID
	})

	return registry
}

// ListPools returns public pool summaries sorted by pool ID.
func (r *Registry) ListPools() []cloudsg.CloudSecurityGroupPoolSummary {
	return append([]cloudsg.CloudSecurityGroupPoolSummary(nil), r.summaries...)
}

// Has reports whether the pool ID is configured.
func (r *Registry) Has(poolID string) bool {
	_, ok := r.pools[poolID]
	return ok
}

// SecurityGroupIDs returns the security group ids of one pool, empty when unknown.
func (r *Registry) SecurityGroupIDs(poolID string) []string {
	pool, ok := r.pools[poolID]
	if !ok {
		return nil
	}

	return append([]string(nil), pool.SecurityGroupIDs...)
}

// Authorize adds missing rules to one pool under a distributed per-pool lock.
func (r *Registry) Authorize(
	nCtx contextx.IContext, poolID string, rules []cloudsg.CloudSecurityGroupRule,
) (err error) {

	pool, err := r.getPool(poolID)
	if err != nil {
		return err
	}

	handler, err := r.vendorHandler(pool.Vendor)
	if err != nil {
		return err
	}

	mutex := r.locker.NewMutex("bk-nodemgr:cloud-security-group:authorize:" + poolID)
	if err := mutex.TryLock(); err != nil {
		return fmt.Errorf("failed to lock cloud security group pool(%s): %w", poolID, err)
	}
	defer func() {
		err = errors.Join(err, mutex.Unlock())
	}()

	if nCtx != nil && nCtx.Err() != nil {
		return fmt.Errorf("cloud security group pool(%s) authorization canceled: %w", poolID, nCtx.Err())
	}

	return handler.Authorize(nCtx, providerPool(pool), r.NormalizeRules(poolID, rules))
}

// Inspect queries live rules of one pool without modifying them.
func (r *Registry) Inspect(
	nCtx contextx.IContext, poolID string,
) ([]cloudsg.CloudSecurityGroupInspection, error) {

	pool, err := r.getPool(poolID)
	if err != nil {
		return nil, err
	}

	handler, err := r.vendorHandler(pool.Vendor)
	if err != nil {
		return nil, err
	}

	return handler.Inspect(nCtx, providerPool(pool))
}

// vendorHandler selects the provider implementation by the pool configured vendor.
func (r *Registry) vendorHandler(vendor cloudsg.CloudSecurityGroupVendor) (tencentcloud.ISecurityGroupHandler, error) {
	handler, ok := r.handlers[vendor]
	if !ok {
		return nil, fmt.Errorf("no security group handler registered for vendor: %s", vendor)
	}

	return handler, nil
}

func (r *Registry) getPool(poolID string) (config.CloudSecurityGroupPool, error) {
	pool, ok := r.pools[poolID]
	if !ok {
		return config.CloudSecurityGroupPool{}, fmt.Errorf("unknown cloud security group pool: %s", poolID)
	}

	return pool, nil
}

// NormalizeRules applies the pool level rule transformations. When the pool enables
// portAll, every rule is authorized with ALL ports regardless of the endpoint ports.
func (r *Registry) NormalizeRules(poolID string, rules []cloudsg.CloudSecurityGroupRule) []cloudsg.CloudSecurityGroupRule {
	pool, ok := r.pools[poolID]
	if !ok || !pool.PortAll {
		return rules
	}

	normalized := make([]cloudsg.CloudSecurityGroupRule, len(rules))
	for idx, rule := range rules {
		rule.Port = "ALL"
		normalized[idx] = rule
	}

	return normalized
}

func providerPool(pool config.CloudSecurityGroupPool) tencentcloud.SecurityGroupPool {
	return tencentcloud.SecurityGroupPool{
		Region:           pool.Region,
		Endpoint:         pool.Endpoint,
		SecretID:         pool.SecretID,
		SecretKey:        pool.SecretKey,
		SecurityGroupIDs: append([]string(nil), pool.SecurityGroupIDs...),
	}
}
