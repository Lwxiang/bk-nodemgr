/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package operinstdata ...
package operinstdata

import (
	"context"
	"os"
	"testing"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/logger"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow/operengine"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testClient ...
func testClient(t *testing.T) Handler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	mongoClient, err := mongo.Connect(
		ctx,
		&options.ClientOptions{
			Hosts: []string{
				os.Getenv("MONGO_ADDRESS"),
			},
			Auth: &options.Credential{
				Username:      os.Getenv("MONGO_USER"),
				Password:      os.Getenv("MONGO_PASSWORD"),
				AuthSource:    os.Getenv("MONGO_AUTH_SOURCE"),
				AuthMechanism: os.Getenv("MONGO_AUTH_MECHANISM"),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return New(mongoClient.Database(os.Getenv("MONGO_DATABASE")), logger.LoggerDefault{})
}

// Test_handler_FindOne ...
func Test_handler_FindOne(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.Background(),
				opts: []OptFn{
					WithOperInstID("operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			got, err := h.FindOne(tt.args.ctx, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOne() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("got = %#v\n", got)

			for name, data := range got.ActionInstDataMap {
				t.Logf("action = %s, data = %#v\n", name, data)
			}
		})
	}
}

// Test_handler_Upsert ...
func Test_handler_Upsert(t *testing.T) {
	type args struct {
		ctx  context.Context
		data *operengine.OperInstData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.Background(),
				data: &operengine.OperInstData{
					OperInstID:  "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
					OperDefName: "",
					ActionNames: []string{"action-1"},
					ActionInstDataMap: map[string]*operengine.ActionInstData{
						"action-1": {
							TriggerID:  "trigger-1",
							OperInstID: "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
							Name:       "action-1",
							Index:      0,
							State:      "success",
							StartedAt:  time.Time{},
							EndedAt:    time.Time{},
							Messages:   nil,
							Content: map[string]any{
								"biz": "1",
							},
						},
					},
					InitContent: map[string]map[string]any{},
				},
			},
			wantErr: false,
		},
		{
			name: "nil init content",
			args: args{
				ctx: context.Background(),
				data: &operengine.OperInstData{
					OperInstID:  "operation-inst-7bd49883-bcc9-4776-80ff-d3d37ca4143f",
					InitContent: nil,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			if err := h.Upsert(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_handler_RefreshActInstDataMsg ...
func Test_handler_RefreshActInstDataMsg(t *testing.T) {
	type args struct {
		ctx  context.Context
		data *operengine.ActionInstData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.Background(),
				data: &operengine.ActionInstData{
					TriggerID:  "trigger-1",
					OperInstID: "a19daad7-aaa5-4a1b-a74c-29e33eeae928",
					Name:       "sync_biz_from_cmdb",
					Index:      0,
					Messages: []operengine.Message{
						{
							Time: time.Time{},
							Text: "success",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nil content",
			args: args{
				ctx: nil,
				data: &operengine.ActionInstData{
					TriggerID:  "",
					OperInstID: "",
					Name:       "",
					Index:      0,
					State:      "",
					StartedAt:  time.Time{},
					EndedAt:    time.Time{},
					StoppedAt:  time.Time{},
					Messages:   nil,
					Content:    nil,
				},
			},
			wantErr: true,
		},
		{
			name: "nil data",
			args: args{
				ctx:  context.Background(),
				data: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testClient(t)
			err := h.RefreshActInstDataMsg(tt.args.ctx, tt.args.data)
			if err != nil {
				t.Logf("RefreshActInstDataMsg() error = %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshActInstDataMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
