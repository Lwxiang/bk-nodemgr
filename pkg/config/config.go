/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package config provides configuration.
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Etcd the config of etcd.
type Etcd struct {
	Endpoints string `yaml:"endpoints" usage:"endpoints of etcd"`
	Cert      string `yaml:"cert" usage:"cert file of etcd"`
	Key       string `yaml:"key" usage:"key file for etcd"`
	Ca        string `yaml:"ca" usage:"ca file for etcd"`
}

// Redis the config of redis.
type Redis struct {
	Host     string `yaml:"host" usage:"host of redis"`
	Port     int    `yaml:"port" usage:"port of redis"`
	Password string `yaml:"password" usage:"password of redis"`
	DB       int    `yaml:"db" usage:"db of redis"`
}

// MongoDB the config of mongodb.
type MongoDB struct {
	Hosts         []string `yaml:"hosts" usage:"hosts list of mongodb"`
	Username      string   `yaml:"username" usage:"user of mongodb"`
	Password      string   `yaml:"password" usage:"password of mongodb"`
	AuthSource    string   `yaml:"auth_source" usage:"auth source of mongodb"`
	AuthMechanism string   `yaml:"auth_mechanism" usage:"auth mechanism of mongodb"`
}

// Log the config of log.
type Log struct {
	Dir          string `yaml:"dir" usage:"log dir of backend server"`
	MaxSizeMB    uint64 `yaml:"max_size_mb" usage:"max size in MBytes of single log file"`
	MaxNum       int    `yaml:"max_num" usage:"max number of log files"`
	Level        string `yaml:"level" usage:"log level of backend server. DEBUG, INFO, WARN, ERROR"`
	ToStdErr     bool   `yaml:"to_stderr" usage:"log to stderr instead of files"`
	AlsoToStdErr bool   `yaml:"also_to_stderr" usage:"log to stderr in addition to files"`
}

// HTTPServer the config of http service.
type HTTPServer struct {
	BindIP string `yaml:"bind_ip" usage:"bind ip of http server"`
	Port   int    `yaml:"port" usage:"port of http server"`
}

// APIGateway the config of api-gateway.
type APIGateway struct {
	AppCode          string `yaml:"app_code" usage:"app code for api-gateway"`
	AppSecret        string `yaml:"app_secret" usage:"app secret for api-gateway"`
	Domain           string `yaml:"domain" usage:"domain of api-gateway"`
	PlatformUsername string `yaml:"platform_username" usage:"platform username for api-gateway"`
}

// CMDB the config of cmdb.
type CMDB struct {
	Environment string `yaml:"environment" usage:"environment of cmdb"`
}

// BackendService the config of backend service.
type BackendService struct {
	APIGateway APIGateway `yaml:"api_gateway" usage:"auth config of backend service"`
	CMDB       CMDB       `yaml:"cmdb" usage:"cmdb config of backend service"`
	HTTPServer HTTPServer `yaml:"http_server" usage:"http server config of backend service"`
	Redis      Redis      `yaml:"redis" usage:"redis config of backend service"`
	MongoDB    MongoDB    `yaml:"mongodb" usage:"mongodb config of backend service"`
	Log        Log        `yaml:"log" usage:"log config of backend service"`
}

// LoadFromFile loads config from file.
func (b *BackendService) LoadFromFile(path string) error {
	configContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(configContent, b); err != nil {
		return err
	}

	return nil
}

// Validate validates the config.
func (b *BackendService) Validate() error {
	// TODO: validate the config
	return nil
}
