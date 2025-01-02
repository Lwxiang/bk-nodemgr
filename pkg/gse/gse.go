/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package gse provides handlers to operate gse API.
package gse

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/apigw"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"github.com/google/uuid"
)

// Handler is the interface for gse handler.
type Handler interface {
	// ListAgentInfo list agent detail information.
	// @param agentIDList given agent id list.
	// @return agentInfoList agent detail information list.
	ListAgentInfo(agentIDList []string) ([]*types.AgentInfo, error)

	// ListAgentState list agent state information. AgentState is a subset of AgentInfo.
	// This method is more efficient than ListAgentInfo.
	// @param agentIDList given agent id list.
	// @return agentStateList agent state information list.
	ListAgentState(agentIDList []string) ([]*types.AgentState, error)

	// ExecuteScript execute script on host.
	// @param endpoints given endpoint list with auth.
	// @param scriptContent given script content.
	// @param timeout given timeout.
	// @return gse-task-id for this execution for further querying.
	ExecuteScript(endpoints []*types.EndpointWithAuth, scriptContent string, timeout time.Duration) (string, error)

	// QueryScriptExecutionResult query script execution result.
	// @param taskID given task id.
	// @param endpoints given endpoint list.
	// @return script result list.
	QueryScriptExecutionResult(taskID string, endpoints []*types.EndpointWithRestrict) ([]*types.ScriptResult, error)

	// TerminateScriptExecution terminate script execution.
	// @param taskID given task id
	// @param endpoints given endpoint list.
	// @return gse-task-id for this operation.
	TerminateScriptExecution(taskID string, endpoints []*types.Endpoint) (string, error)

	// TransferFile transfer files from source to targets.
	// @param opts given options.
	// @param transfers given transfer details.
	// @return gse-task-id for this transferring.
	TransferFile(opts *types.TransferOptions, transfers []*types.TransferDetail) (string, error)

	// QueryFileTransmissionResult query file transmission result.
	// @param taskID given task id.
	// @param endpoints given endpoint list.
	// @return file transmission result list.
	QueryFileTransmissionResult(taskID string, endpoints []*types.Endpoint) ([]*types.TransferResult, error)

	// TerminateFileTransmission terminate file transmission.
	// @param taskID given task id.
	// @param endpoints given endpoint list.
	// @return gse-task-id for this operation.
	TerminateFileTransmission(taskID string, endpoints []*types.Endpoint) (string, error)
}

// Config defines the config for gse handler.
type Config struct {
	Environment string
}

// NewHandler creates a new gse handler.
func NewHandler(config *Config, apigwCli apigw.Client) Handler {
	return &handler{
		config:   config,
		apigwCli: apigwCli,
	}
}

type handler struct {
	config   *Config
	apigwCli apigw.Client
}

// ListAgentInfo list agent detail information.
func (hld *handler) ListAgentInfo(agentIDList []string) ([]*types.AgentInfo, error) {
	body, err := json.Marshal(&ReqListAgentInfo{
		AgentIDList: agentIDList,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/cluster/list_agent_info", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespListAgentInfo{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, respData, err)
	}

	data := make([]*types.AgentInfo, 0, len(resp.Data))
	for _, info := range resp.Data {
		data = append(data, &types.AgentInfo{
			AgentState: types.AgentState{
				AgentID:        info.BKAgentID,
				CloudID:        info.BKCloudID,
				Version:        info.Version,
				StatusCode:     types.AgentStatusCode(info.StatusCode),
				LastStatusCode: types.AgentStatusCode(info.LastStatusCode),
				ReportTime:     info.ReportTime,
			},
			HostIP:        info.BKHostIP,
			OSType:        info.BKOSType,
			ParentIP:      info.ParentIP,
			ParentPort:    info.ParentPort,
			CPURate:       info.CPURate,
			MemRate:       info.MemRate,
			StartTime:     info.StartTime,
			LastWorkTime:  info.LastWorkTime,
			ConnCycleTime: info.ConnCycleTime,
			Status:        info.Status,
			LastStatus:    info.LastStatus,
			RunMode:       info.RunMode,
			Remark:        info.Remark,
		})
	}

	return data, nil
}

// ListAgentState list agent state.
func (hld *handler) ListAgentState(agentIDList []string) ([]*types.AgentState, error) {
	body, err := json.Marshal(&ReqListAgentState{
		AgentIDList: agentIDList,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/cluster/list_agent_state", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespListAgentState{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, respData, err)
	}

	data := make([]*types.AgentState, 0, len(resp.Data))
	for _, info := range resp.Data {
		data = append(data, &types.AgentState{
			AgentID:        info.BKAgentID,
			CloudID:        info.BKCloudID,
			Version:        info.Version,
			StatusCode:     types.AgentStatusCode(info.StatusCode),
			LastStatusCode: types.AgentStatusCode(info.LastStatusCode),
			ReportTime:     info.ReportTime,
		})
	}

	return data, nil
}

// ExecuteScript execute script on hosts.
func (hld *handler) ExecuteScript(
	endpoints []*types.EndpointWithAuth, scriptContent string, timeout time.Duration) (string, error) {

	eps := make([]*EndpointWithAuth, 0, len(endpoints))
	for _, endpoint := range endpoints {
		eps = append(eps, &EndpointWithAuth{
			Endpoint: Endpoint{
				BKAgentID:     endpoint.AgentID,
				BKContainerID: endpoint.ContainerID,
			},
			User:     endpoint.User,
			Password: endpoint.Password,
		})
	}

	scriptName := fmt.Sprintf("bk_gse_script_nodeman_%s.sh", uuid.New().String())
	storedDir := "/tmp/bknodeman/"

	body, err := json.Marshal(&ReqExecuteScript{
		Endpoints: eps,
		Scripts: []*ScriptDetail{
			{
				Name:      scriptName,
				StoredDir: storedDir,
				Content:   scriptContent,
			},
		},
		Atomics: []*ScriptAtomicTask{
			{
				ID:         0,
				Command:    filepath.Join(storedDir, scriptName),
				TimeoutSec: int(timeout.Seconds()),
			},
		},
		Relations: make([]*ScriptAtomicRelation, 0),
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/task/extensions/async_execute_script", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespExecuteScript{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, respData, err)
	}

	return resp.Data.Result.TaskID, nil
}

// QueryScriptExecutionResult query script execution result.
func (hld *handler) QueryScriptExecutionResult(
	taskID string, endpoints []*types.EndpointWithRestrict) ([]*types.ScriptResult, error) {

	conditions := make([]*ScriptEndpointCondition, 0, len(endpoints))
	for _, endpoint := range endpoints {
		conditions = append(conditions, &ScriptEndpointCondition{
			Endpoint: Endpoint{
				BKAgentID:     endpoint.AgentID,
				BKContainerID: endpoint.ContainerID,
			},
			Atomics: []*ScriptAtomicTaskCondition{
				{
					ID:     0,
					Offset: endpoint.Offset,
					Limit:  endpoint.Limit,
				},
			},
		})
	}

	body, err := json.Marshal(&ReqQueryScriptExecutionResult{
		TaskID:     taskID,
		AgentTasks: conditions,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/task/extensions/get_execute_script_result", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespQueryScriptExecutionResult{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, respData, err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("failed to query script execution result. result-code(%d), result-msg(%s)",
			resp.Code, resp.Message)
	}

	result := make([]*types.ScriptResult, 0, len(resp.Data.Result))
	for _, rst := range resp.Data.Result {
		result = append(result, &types.ScriptResult{
			Endpoint: types.Endpoint{
				AgentID:     rst.BKAgentID,
				ContainerID: rst.BKContainerID,
			},
			Status:       types.ScriptStatus(rst.Status),
			ErrorCode:    rst.ErrorCode,
			ErrorMessage: rst.ErrorMessage,
			StartTime:    time.UnixMilli(rst.StartTime),
			EndTime:      time.UnixMilli(rst.EndTime),
			ExitCode:     rst.ExitCode,
			ScreenLog:    rst.ScreenLog,
		})
	}

	return result, nil
}

// TerminateExecution terminate execution.
func (hld *handler) TerminateScriptExecution(taskID string, endpoints []*types.Endpoint) (string, error) {
	eps := make([]*Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		eps = append(eps, &Endpoint{
			BKAgentID:     endpoint.AgentID,
			BKContainerID: endpoint.ContainerID,
		})
	}

	body, err := json.Marshal(&ReqTerminateScriptExecution{
		TaskID:    taskID,
		Endpoints: eps,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/task/extensions/async_terminate_execute_script", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespTerminateScriptExecution{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, respData, err)
	}

	return resp.Data.Result.TaskID, nil
}

// TransferFile transfer file.
func (hld *handler) TransferFile(opts *types.TransferOptions, transfers []*types.TransferDetail) (string, error) {
	tasks := make([]*TransferDetail, 0, len(transfers))
	for _, transfer := range transfers {
		targetEndpoints := make([]*EndpointWithAuth, 0, len(transfer.Target.Endpoints))
		for _, endpoint := range transfer.Target.Endpoints {
			targetEndpoints = append(targetEndpoints, &EndpointWithAuth{
				Endpoint: Endpoint{
					BKAgentID:     endpoint.AgentID,
					BKContainerID: endpoint.ContainerID,
				},
				User: endpoint.User,
			})
		}

		tasks = append(tasks, &TransferDetail{
			Source: &TransferSource{
				FileName:  transfer.Source.FileName,
				StoredDir: transfer.Source.StoredDir,
				Endpoint: EndpointWithAuth{
					Endpoint: Endpoint{
						BKAgentID:     transfer.Source.Endpoint.AgentID,
						BKContainerID: transfer.Source.Endpoint.ContainerID,
					},
					User: transfer.Source.Endpoint.User,
				},
			},
			Target: &TransferTarget{
				StoredDir: transfer.Target.StoredDir,
				Endpoints: targetEndpoints,
			},
		})
	}

	body, err := json.Marshal(&ReqTransferFile{
		TimeoutSec:    uint(opts.Timeout.Seconds()),
		AutoMkdir:     opts.AutoMkdir,
		UploadSpeed:   opts.UploadSpeedMBPerSec,
		DownloadSpeed: opts.DownloadSpeedMBPerSec,
		Tasks:         tasks,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/task/extensions/async_transfer_file", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespTransferFile{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return "", fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	return resp.Data.Result.TaskID, nil
}

// QueryFileTransmissionResult query file transmission result.
func (hld *handler) QueryFileTransmissionResult(
	taskID string, endpoints []*types.Endpoint) ([]*types.TransferResult, error) {

	eps := make([]*Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		eps = append(eps, &Endpoint{
			BKAgentID:     endpoint.AgentID,
			BKContainerID: endpoint.ContainerID,
		})
	}

	body, err := json.Marshal(&ReqQueryFileTransmissionResult{
		TaskID:    taskID,
		Endpoints: eps,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/task/extensions/get_transfer_file_result", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespQueryFileTransmissionResult{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("failed to query file tranmission result. result-code(%d), result-msg(%s)",
			resp.Code, resp.Message)
	}

	result := make([]*types.TransferResult, 0, len(resp.Data.Result))
	for _, rst := range resp.Data.Result {
		result = append(result, &types.TransferResult{
			Source: types.Endpoint{
				AgentID:     rst.Content.SourceAgentID,
				ContainerID: rst.Content.SourceContainerID,
			},
			Target: types.Endpoint{
				AgentID:     rst.Content.DestinationAgentID,
				ContainerID: rst.Content.DestinationContainerID,
			},
			Mode:           types.TransferMode(rst.Content.Mode),
			Progress:       rst.Content.Progress,
			SpeedKBPerSec:  rst.Content.Speed,
			SizeBytes:      rst.Content.Size,
			SourceDir:      rst.Content.SourceFileDir,
			SourceFileName: rst.Content.SourceFileName,
			TargetDir:      rst.Content.DestFileDir,
			TargetFileName: rst.Content.DestFileName,
			ErrorCode:      rst.ErrorCode,
			ErrorMessage:   rst.ErrorMessage,
			StatusCode:     types.TransferStatus(rst.Content.Status),
			StatusMessage:  rst.Content.StatusInfo,
			StartTime:      time.UnixMilli(rst.Content.StartTime),
			EndTime:        time.UnixMilli(rst.Content.EndTime),
		})
	}

	return result, nil
}

func (hld *handler) TerminateFileTransmission(taskID string, endpoints []*types.Endpoint) (string, error) {
	eps := make([]*Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		eps = append(eps, &Endpoint{
			BKAgentID:     endpoint.AgentID,
			BKContainerID: endpoint.ContainerID,
		})
	}

	body, err := json.Marshal(&ReqTerminateFileTransmission{
		TaskID:    taskID,
		Endpoints: eps,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("api/bk-gse/%s/api/v2/task/extensions/async_terminate_transfer_file", hld.config.Environment)
	code, respData, err := hld.apigwCli.Post(nil, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to do post(%s). http-code(%d), resp(%s), err: %v", url, code, respData, err)
	}

	resp := &RespTerminateFileTransmission{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body. post(%s), data(%s), err: %v", url, respData, err)
	}

	return resp.Data.Result.TaskID, nil
}
