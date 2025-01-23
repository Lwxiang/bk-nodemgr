/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operengine ...
package operengine

import "testing"

// Test_evaluateActionInstanceState ...
func Test_evaluateActionInstanceState(t *testing.T) {
	type args struct {
		data *ActionInstData
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "nil data",
			args: args{
				data: nil,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStateSuccess,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "skipped",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStateSkipped,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "failed",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStateFailed,
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "timeout",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStateTimeout,
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "terminated",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStateTerminated,
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "running",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStateRunning,
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "pending",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStatePending,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "unknown",
			args: args{
				data: &ActionInstData{
					State: ActionInstanceStateUnknown,
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateActionInstanceState(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateActionInstanceState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("evaluateActionInstanceState() got = %v, want %v", got, tt.want)
			}
		})
	}
}
