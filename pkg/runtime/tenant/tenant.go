/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tenant ...
package tenant

import (
	"context"
	"errors"
	"strings"
	"sync"
)

type Mode string

const (
	ModeSingle   Mode = "single"
	ModeMultiple Mode = "multiple"
)

// instance tenant mode.
var instance = struct {
	mode Mode
	once sync.Once
}{
	mode: ModeSingle,
	once: sync.Once{},
}

// SetMode tenant mode only can be set once.
func SetMode(mode Mode) {
	instance.once.Do(func() {
		instance.mode = mode
	})
}

// GetMode get tenant mode.
func GetMode() Mode {
	return instance.mode
}

const (
	// idKey tenant id key stored in context.
	idKey = "nodeman_tenant_id"

	// SingleModeTenantID tenant id for single mode.
	SingleModeTenantID = "single"
)

// GetID get tenant id from context.
func GetID(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("context is nil")
	}

	tenantID, ok := ctx.Value(idKey).(string)
	if !ok {
		return "", errors.New("tenant_id not found in context")
	}

	if tenantID == SingleModeTenantID {
		return tenantID, nil
	}

	if err := validate(tenantID); err != nil {
		return "", err
	}

	return tenantID, nil
}

func validate(tenantID string) error {
	tenantID = strings.ReplaceAll(tenantID, " ", "")
	if len(tenantID) == 0 {
		return errors.New("tenant_id is empty")
	}

	// TODO: support strict tenant id validation

	return nil
}

// SetID set tenant id to context.
func SetID(ctx context.Context, tenantID string) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}

	if instance.mode == ModeSingle {
		return context.WithValue(ctx, idKey, SingleModeTenantID), nil
	}

	if err := validate(tenantID); err != nil {
		return nil, err
	}

	return context.WithValue(ctx, idKey, tenantID), nil
}
