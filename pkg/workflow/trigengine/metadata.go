/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package trigengine ...
package trigengine

import (
	"errors"
	"time"
)

// MetadataOnce will store the metadata of a trigger
type MetadataOnce struct {
}

// Validate ...
func (m *MetadataOnce) Validate() error {
	return nil
}

// MetadataPeriodic will store the metadata of a trigger
type MetadataPeriodic struct {
	IntervalSecond int
}

// Validate ...
func (m *MetadataPeriodic) Validate() error {
	if m.IntervalSecond <= 0 {
		return errors.New("interval second must be greater than 0")
	}

	return nil
}

// GetInterval get interval
func (m *MetadataPeriodic) GetInterval() time.Duration {
	return time.Duration(m.IntervalSecond) * time.Second
}

// MetadataOrdered will store the metadata of a trigger
type MetadataOrdered struct {
}

// Validate validate
func (m *MetadataOrdered) Validate() error {
	return nil
}
