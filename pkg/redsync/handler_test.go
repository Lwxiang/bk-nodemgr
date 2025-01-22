/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package redsync ...
package redsync

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/joho/godotenv"
	goredislib "github.com/redis/go-redis/v9"
)

// testClient ...
func testClient(t *testing.T) Handler {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatal(err)
	}

	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		t.Fatal(err)
	}

	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})

	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		t.Fatal(err)
	}

	return New(redisClient)
}

// Test_mutex_Lock ...
func Test_mutex_Lock(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		isLock   bool
		isUnlock bool
		isExpire bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			isLock:   false,
			isUnlock: false,
			isExpire: false,
			wantErr:  false,
		},
		{
			name: "already_lock",
			args: args{
				ctx: context.Background(),
			},
			wantErr:  true,
			isLock:   true,
			isUnlock: false,
			isExpire: false,
		},
		{
			name: "already_expire, but auto extend",
			args: args{
				ctx: context.Background(),
			},
			wantErr:  true,
			isLock:   true,
			isUnlock: false,
			isExpire: true,
		},
		{
			name: "already_unlock",
			args: args{
				ctx: context.Background(),
			},
			wantErr:  false,
			isLock:   true,
			isUnlock: true,
			isExpire: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testClient(t).NewMutex(tt.name)

			if tt.isLock {
				if err := m.TryLock(); err != nil {
					t.Errorf("Lock() error = %v", err)
				}
			}

			if tt.isUnlock {
				if result, err := m.Unlock(); (err != nil) || result != true {
					t.Errorf("Unlock() error = %v, wantUnlock %v", err, tt.isUnlock)
				}
			}

			if tt.isExpire {
				// default redsync expire time is 8s, inorder to test expire, sleep 10s.
				time.Sleep(time.Second * 10)
			}

			if err := m.TryLock(); (err != nil) != tt.wantErr {
				t.Errorf("Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_mutex_Name ...
func Test_mutex_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "success",
			want: "success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testClient(t).NewMutex(tt.name)
			if got := m.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_mutex_Unlock ...
func Test_mutex_Unlock(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		args     args
		want     bool
		wantErr  bool
		isLock   bool
		isUnlock bool
		isExpire bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			want:     true,
			wantErr:  false,
			isLock:   true,
			isUnlock: false,
			isExpire: false,
		},
		{
			name:     "no_lock",
			args:     args{},
			want:     false,
			wantErr:  true,
			isLock:   false,
			isUnlock: false,
			isExpire: false,
		},
		{
			name: "already_unlock",
			args: args{
				ctx: context.Background(),
			},
			want:     false,
			wantErr:  true,
			isLock:   true,
			isUnlock: true,
			isExpire: false,
		},
		{
			name: "expire",
			args: args{
				ctx: context.Background(),
			},
			want:     false,
			wantErr:  true,
			isLock:   true,
			isUnlock: false,
			isExpire: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testClient(t).NewMutex(tt.name)

			if tt.isLock {
				if err := m.TryLock(); err != nil {
					t.Errorf("Lock() error = %v", err)
				}
			}

			if tt.isUnlock {
				if result, err := m.Unlock(); (err != nil) || result != true {
					t.Errorf("Unlock() error = %v, wantUnlock %v", err, tt.isUnlock)
				}
			}

			if tt.isExpire {
				// default redsync expire time is 8s, inorder to test expire, sleep 10s.
				time.Sleep(time.Second * 10)
			}

			got, err := m.Unlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("Unlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Unlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}
