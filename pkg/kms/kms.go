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

// Package kms handles the secret from kms.
package kms

import (
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-kms-sdk/go/decrypt"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/kms/envelope"
)

// LoadParam the load param.
type LoadParam struct {
	// global envelope and private-key.
	EnvelopeGlobal   string
	PrivateKeyGlobal string

	// module envelope and private-key.
	EnvelopeModule   string
	PrivateKeyModule string
}

// LoadKMSApplicationSecret load the kms application secret.
func LoadKMSApplicationSecret(param LoadParam) (*envelope.KMSGlobalEnvelope, *envelope.KMSApplicationEnvelope, error) {
	dataGlobal, err := loadKMSGlobalSecret(param.EnvelopeGlobal, param.PrivateKeyGlobal)
	if err != nil {
		return nil, nil, err
	}

	dataModule, err := loadKMSApplicationSecret(param.EnvelopeModule, param.PrivateKeyModule)
	if err != nil {
		return nil, nil, err
	}

	return dataGlobal, dataModule, nil
}

// LoadKMSBackendSecret load the kms backend secret.
func LoadKMSBackendSecret(param LoadParam) (*envelope.KMSGlobalEnvelope, *envelope.KMSBackendEnvelope, error) {
	dataGlobal, err := loadKMSGlobalSecret(param.EnvelopeGlobal, param.PrivateKeyGlobal)
	if err != nil {
		return nil, nil, err
	}

	dataModule, err := loadKMSBackendSecret(param.EnvelopeModule, param.PrivateKeyModule)
	if err != nil {
		return nil, nil, err
	}

	return dataGlobal, dataModule, nil
}

// LoadKMSFileSecret load the kms file secret.
func LoadKMSFileSecret(param LoadParam) (*envelope.KMSGlobalEnvelope, *envelope.KMSFileEnvelope, error) {
	dataGlobal, err := loadKMSGlobalSecret(param.EnvelopeGlobal, param.PrivateKeyGlobal)
	if err != nil {
		return nil, nil, err
	}

	dataModule, err := loadKMSFileSecret(param.EnvelopeModule, param.PrivateKeyModule)
	if err != nil {
		return nil, nil, err
	}

	return dataGlobal, dataModule, nil
}

func loadKMSGlobalSecret(envelopeGlobal, key string) (*envelope.KMSGlobalEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeGlobal, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt global secret: %w", err)
	}

	var data envelope.KMSGlobalEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal global secret: %w", err)
	}

	return &data, nil
}

func loadKMSApplicationSecret(envelopeModule, key string) (*envelope.KMSApplicationEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeModule, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt application secret: %w", err)
	}

	var data envelope.KMSApplicationEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal application secret: %w", err)
	}

	return &data, nil
}

func loadKMSBackendSecret(envelopeModule, key string) (*envelope.KMSBackendEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeModule, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt backend secret: %w", err)
	}

	var data envelope.KMSBackendEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal backend secret: %w", err)
	}

	return &data, nil
}

func loadKMSFileSecret(envelopeModule, key string) (*envelope.KMSFileEnvelope, error) {
	plainText, err := decrypt.Decrypt(envelopeModule, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file secret: %w", err)
	}

	var data envelope.KMSFileEnvelope
	err = json.Unmarshal([]byte(plainText), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file secret: %w", err)
	}

	return &data, nil
}
