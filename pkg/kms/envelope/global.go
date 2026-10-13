/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

// Package envelope defines the decrypted KMS envelope structures.
package envelope

// KMSGlobalEnvelope defines kms global envelope.
type KMSGlobalEnvelope struct {
	BKAppIDSecret struct {
		Default KMSBKAppIDSecret `json:"default"`
	} `json:"bk_app_id_secret"`

	Etcd struct {
		Default KMSEtcd `json:"default"`
	} `json:"bk_etcd_secret"`

	Redis struct {
		Default KMSRedis `json:"default"`
	} `json:"redis"`

	MongoDB struct {
		Default KMSMongoDB `json:"default"`
	} `json:"mongodb"`

	JwtServerConfig KMSJwtServerConfig `json:"jwt_server_config"`
}

// KMSBKAppIDSecret defines app code and app secret in kms.
type KMSBKAppIDSecret struct {
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
}

// KMSEtcd defines etcd config in kms.
type KMSEtcd struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	TLSCABase64   string `json:"tls_ca_base64"`
	TLSCertBase64 string `json:"tls_cert_base64"`
	TLSKeyBase64  string `json:"tls_key_base64"`
	TLSPassword   string `json:"tls_password"`
}

// KMSRedis defines redis config in kms.
type KMSRedis struct {
	Password string `json:"password"`
}

// KMSMongoDB defines mongodb config in kms.
type KMSMongoDB struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// KMSJwtServerConfig defines jwt server config in kms.
type KMSJwtServerConfig struct {
	ApplicationInfo  KMSJwtServerConfigItem `json:"application_info"`
	ApplicationAdmin KMSJwtServerConfigItem `json:"application_admin"`
	ApplicationBasic KMSJwtServerConfigItem `json:"application_basic"`
	BackendInfo      KMSJwtServerConfigItem `json:"backend_info"`
	BackendAdmin     KMSJwtServerConfigItem `json:"backend_admin"`
	BackendBasic     KMSJwtServerConfigItem `json:"backend_basic"`
	BackendCallback  KMSJwtServerConfigItem `json:"backend_callback"`
	BackendProxy     KMSJwtServerConfigItem `json:"backend_proxy"`
	FileInfo         KMSJwtServerConfigItem `json:"file_info"`
	FileAdmin        KMSJwtServerConfigItem `json:"file_admin"`
	FileBasic        KMSJwtServerConfigItem `json:"file_basic"`
	FileDownload     KMSJwtServerConfigItem `json:"file_download"`
	FileExport       KMSJwtServerConfigItem `json:"file_export"`
}

// KMSJwtServerConfigItem defines jwt server config item in kms.
type KMSJwtServerConfigItem struct {
	SymmetricKey       string `json:"symmetric_key"`
	PublicKeyPEMBase64 string `json:"public_key_pem_base64"`
}
