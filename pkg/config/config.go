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
	"fmt"
	"os"

	"git.woa.com/bk-gse/bk-nodeman/pkg/envx"
	"gopkg.in/yaml.v2"
)

const (
	// backend service config default values.
	defaultBackendRunMode      = RunModeRelease
	defaultBackendHTTPBindIP   = "127.0.0.1"
	defaultBackendHTTPPort     = 8000
	defaultBackendLogDir       = "/bk-nodeman/log/"
	defaultBackendLogMaxNum    = 10
	defaultBackendLogMaxSizeMB = 200
	defaultBackendLogLevel     = "INFO"

	// saas service config default values.
	defaultSaasRunMode       = RunModeRelease
	defaultSaasAPIGwUser     = "admin"
	defaultSaasHTTPBindIP    = "127.0.0.1"
	defaultSaasHTTPPort      = 5000
	defaultSaasHTTPStaticDir = "/bk-nodeman/static/"
	defaultSaasLogDir        = "/bk-nodeman/log/"
	defaultSaasLogMaxNum     = 10
	defaultSaasLogMaxSizeMB  = 200
	defaultSaasLogLevel      = "INFO"
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

// Validate configures the config.
func (conf Redis) Validate() error {
	return nil
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
	MaxSizeMB    int    `yaml:"max_size_mb" usage:"max size in MBytes of single log file"`
	MaxNum       int    `yaml:"max_num" usage:"max number of log files"`
	Level        string `yaml:"level" usage:"log level of backend server. DEBUG, INFO, WARN, ERROR"`
	ToStdErr     bool   `yaml:"to_stderr" usage:"log to stderr instead of files"`
	AlsoToStdErr bool   `yaml:"also_to_stderr" usage:"log to stderr in addition to files"`
}

// HTTPServer the config of http service.
type HTTPServer struct {
	BindIP    string `yaml:"bind_ip"`
	Port      int    `yaml:"port"`
	StaticDir string `yaml:"static_dir"`
}

// AdminServer the config of admin service.
type AdminServer struct {
	BindIP    string `yaml:"bind_ip"`
	Port      int    `yaml:"port"`
	StaticDir string `yaml:"static_dir"`
}

// APIGateway the config of api-gateway.
type APIGateway struct {
	// Endpoints is a seed list of host:port addresses of api gateway nodes.
	Endpoints []string `yaml:"endpoints"`
	// AppCode is the BlueKing app code of nodeman to request api gateway.
	AppCode string `yaml:"appCode"`
	// AppSecret is the BlueKing app secret of nodeman to request api gateway.
	AppSecret string `yaml:"appSecret"`
	// User is the BlueKing user of nodeman to request api gateway.
	User string `yaml:"user"`
	// AuthMode is the BlueKing api authentication mode.
	AuthMode string `yaml:"authMode"`
	// BkTicket is the BlueKing access ticket of nodeman to request api gateway.
	BkTicket string `yaml:"bkTicket"`
	// BkToken is the BlueKing user token of nodeman to request api gateway.
	BkToken string `yaml:"bkToken"`
	// AccessToken is the BlueKing access token of nodeman to request api gateway.
	AccessToken string `yaml:"accessToken"`
	// TLS defines the tls config of api-gateway.
	TLS TLSConfig `yaml:"tls" usage:"tls config of api-gateway"`
}

// CMDB the config of cmdb.
type CMDB struct {
	TenantID   string `yaml:"tenant_id" usage:"tenant id of cmdb"`
	APIGateway `yaml:",inline" usage:"api-gateway config of cmdb"`
}

// TLSConfig defines tls related options.
type TLSConfig struct {
	// Server should be accessed without verifying the TLS certificate.
	// For testing only.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
	// Server requires TLS client certificate authentication
	CertFile string `yaml:"certFile"`
	// Server requires TLS client certificate authentication
	KeyFile string `yaml:"keyFile"`
	// Trusted root certificates for server
	CAFile string `yaml:"caFile"`
	// the password to decrypt the certificate
	Password string `yaml:"password"`
}

// Workflow the config of workflow.
type Workflow struct {
	WorkerNum int   `yaml:"worker_num" usage:"worker num of workflow"`
	Redis     Redis `yaml:"redis" usage:"redis config of backend service"`
}

// Validate validates the config.
func (conf Workflow) Validate() error {
	if conf.WorkerNum <= 0 {
		return fmt.Errorf("worker num must be greater than 0, worker-num(%d)", conf.WorkerNum)
	}

	if err := conf.Redis.Validate(); err != nil {
		return err
	}

	return nil
}

// RunMode the run mode of service.
type RunMode string

const (
	// RunModeRelease release mode.
	RunModeRelease RunMode = "release"

	// RunModeDebug debug mode.
	RunModeDebug RunMode = "debug"
)

// BackendService the config of backend service.
type BackendService struct {
	RunMode    RunMode    `yaml:"run_mode" usage:"run mode of service"`
	TenantMode string     `yaml:"tenant_mode" usage:"tenant mode of service"`
	CMDB       CMDB       `yaml:"cmdb" usage:"cmdb config of backend service"`
	HTTPServer HTTPServer `yaml:"http_server" usage:"http server config of backend service"`
	Redis      Redis      `yaml:"redis" usage:"redis config of backend service"`
	MongoDB    MongoDB    `yaml:"mongodb" usage:"mongodb config of backend service"`
	Log        Log        `yaml:"log" usage:"log config of backend service"`
	Workflow   Workflow   `yaml:"workflow" usage:"workflow config of backend service"`
	AdminServer AdminServer `yaml:"admin_server" usage:"admin server config of backend service"`
}

// NewBackendService generates a new BackendService with default values.
func NewBackendService() *BackendService {
	return &BackendService{
		RunMode: defaultBackendRunMode,
		HTTPServer: HTTPServer{
			BindIP: defaultBackendHTTPBindIP,
			Port:   defaultBackendHTTPPort,
		},
		Log: Log{
			Dir:       defaultBackendLogDir,
			MaxSizeMB: defaultBackendLogMaxSizeMB,
			MaxNum:    defaultBackendLogMaxNum,
			Level:     defaultBackendLogLevel,
		},
	}
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
	if err := b.Workflow.Validate(); err != nil {
		return err
	}

	return nil
}

// SaasService the config of saas service.
type SaasService struct {
	RunMode    RunMode    `yaml:"mode" usage:"run mode of service"`
	APIGateway APIGateway `yaml:"api_gateway" usage:"auth config of backend service"`
	HTTPServer HTTPServer `yaml:"http_server" usage:"http server config of backend service"`
	Log        Log        `yaml:"log" usage:"log config of backend service"`
}

// NewSaasService generatea a new SaasService with default values.
func NewSaasService() *SaasService {
	return &SaasService{
		RunMode: defaultSaasRunMode,
		APIGateway: APIGateway{
			User: defaultSaasAPIGwUser,
		},
		HTTPServer: HTTPServer{
			BindIP:    defaultSaasHTTPBindIP,
			Port:      defaultSaasHTTPPort,
			StaticDir: defaultSaasHTTPStaticDir,
		},
		Log: Log{
			Dir:       defaultSaasLogDir,
			MaxSizeMB: defaultSaasLogMaxSizeMB,
			MaxNum:    defaultSaasLogMaxNum,
			Level:     defaultSaasLogLevel,
		},
	}
}

// Load loads config from file or environment variables.
func (svc *SaasService) Load(filePath string) error {
	// default options.
	svc.RunMode = RunModeRelease

	if filePath == "" {
		return svc.LoadFromEnv()
	}

	return svc.LoadFromFile(filePath)
}

// LoadFromEnv loads config from environment variables.
func (svc *SaasService) LoadFromEnv() error {
	// run mode.
	var runMode string
	if envx.LoadString("NODEMAN_MODE", &runMode) {
		svc.RunMode = RunMode(runMode)
	}

	// api_gateway.
	if err := envx.MustLoadString("BKPAAS_APP_ID", &svc.APIGateway.AppCode); err != nil {
		return err
	}
	if err := envx.MustLoadString("BKPAAS_APP_SECRET", &svc.APIGateway.AppSecret); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_APIGW_USER", &svc.APIGateway.User)
	_ = envx.LoadString("NODEMAN_APIGW_AUTH_MODE", &svc.APIGateway.AuthMode)
	_ = envx.LoadString("NODEMAN_APIGW_BK_TICKET", &svc.APIGateway.BkTicket)
	_ = envx.LoadString("NODEMAN_APIGW_BK_TOKEN", &svc.APIGateway.BkToken)
	_ = envx.LoadString("NODEMAN_APIGW_ACCESS_TOKEN", &svc.APIGateway.AccessToken)
	if _, err := envx.LoadBool("NODEMAN_APIGW_TLS_SKIP_VERIFY", &svc.APIGateway.TLS.InsecureSkipVerify); err != nil {
		return err
	}
	_ = envx.LoadString("NODEMAN_APIGW_TLS_CERT", &svc.APIGateway.TLS.CertFile)
	_ = envx.LoadString("NODEMAN_APIGW_TLS_KEY", &svc.APIGateway.TLS.KeyFile)
	_ = envx.LoadString("NODEMAN_APIGW_TLS_CA", &svc.APIGateway.TLS.CAFile)
	_ = envx.LoadString("NODEMAN_APIGW_TLS_PASSWORD", &svc.APIGateway.TLS.Password)

	// http_server.
	_ = envx.LoadString("NODEMAN_HTTPSVR_BIND_IP", &svc.HTTPServer.BindIP)
	if _, err := envx.LoadInt("NODEMAN_HTTPSVR_PORT", &svc.HTTPServer.Port); err != nil {
		return err
	}

	// log.
	_ = envx.LoadString("NODEMAN_LOG_DIR", &svc.Log.Dir)
	_ = envx.LoadString("NODEMAN_LOG_LEVEL", &svc.Log.Level)
	if _, err := envx.LoadInt("NODEMAN_LOG_MAX_NUM", &svc.Log.MaxNum); err != nil {
		return err
	}
	if _, err := envx.LoadInt("NODEMAN_LOG_MAX_SIZE_MB", &svc.Log.MaxSizeMB); err != nil {
		return err
	}

	return nil
}

// LoadFromFile loads config from file.
func (svc *SaasService) LoadFromFile(path string) error {
	configContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(configContent, svc); err != nil {
		return err
	}

	return nil
}

// Validate validates the config.
func (svc *SaasService) Validate() error {
	// TODO: validate the config
	return nil
}

// EnvGet read env, supports default value.
func EnvGet(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

// EnvMustGet read env, panic if not set.
func EnvMustGet(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	panic(fmt.Sprintf("required environment variable %s unset", key))
}
