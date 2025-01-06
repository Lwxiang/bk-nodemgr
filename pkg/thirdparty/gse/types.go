/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package gse

// AgentInfo describes the agent information from gse.
type AgentInfo struct {
	BKAgentID      string  `json:"bk_agent_id"`
	BKCloudID      int     `json:"bk_cloud_id"`
	BKHostIP       string  `json:"bk_host_ip"`
	BKOSType       string  `json:"bk_os_type"`
	ReportTime     uint64  `json:"report_time"`
	ParentIP       string  `json:"parent_ip"`
	ParentPort     uint    `json:"parent_port"`
	Version        string  `json:"version"`
	CPURate        float32 `json:"cpu_rate"`
	MemRate        float32 `json:"mem_rate"`
	StartTime      uint64  `json:"start_time"`
	LastWorkTime   uint64  `json:"last_work_time"`
	ConnCycleTime  string  `json:"conn_cycle_time"`
	RunMode        int     `json:"run_mode"`
	StatusCode     int     `json:"status_code"`
	Status         string  `json:"status"`
	LastStatusCode int     `json:"last_status_code"`
	LastStatus     string  `json:"last_status"`
	Remark         string  `json:"remark"`
}

// AgentState describes the agent state. It is a subset of AgentInfo.
type AgentState struct {
	BKAgentID      string `json:"bk_agent_id"`
	BKCloudID      int    `json:"bk_cloud_id"`
	Version        string `json:"version"`
	ReportTime     uint64 `json:"report_time"`
	StatusCode     int    `json:"status_code"`
	Status         string `json:"status"`
	LastStatusCode int    `json:"last_status_code"`
	LastStatus     string `json:"last_status"`
}

// Endpoint describes the basic endpoint information.
type Endpoint struct {
	BKAgentID     string `json:"bk_agent_id"`
	BKContainerID string `json:"bk_container_id"`
}

// EndpointWithAuth describes the endpoint with auth.
type EndpointWithAuth struct {
	Endpoint

	User     string `json:"user"`
	Password string `json:"pwd"`
}

// RespCommon describes the common response.
type RespCommon struct {
	Code    int    `json:"code"`
	Message string `json:"messaage"`
}

// ReqListAgentInfo describes the request data of list_agent_info.
type ReqListAgentInfo struct {
	AgentIDList []string `json:"agent_id_list"`
}

// RespListAgentInfo describes the response data of list_agent_info.
type RespListAgentInfo struct {
	RespCommon

	Data []*AgentInfo `json:"data"`
}

// ReqListAgentState describes the request data of list_agent_state.
type ReqListAgentState struct {
	AgentIDList []string `json:"agent_id_list"`
}

// RespListAgentState describe the response data of list_agent_state.
type RespListAgentState struct {
	RespCommon

	Data []*AgentState `json:"data"`
}

// ScriptDetail describes the script detail.
type ScriptDetail struct {
	Name      string `json:"script_name"`
	StoredDir string `json:"script_store_dir"`
	Content   string `json:"script_content"`
}

// ScriptAtomicTask describes the atomic task of script.
type ScriptAtomicTask struct {
	ID         int    `json:"atomic_task_id"`
	Command    string `json:"command"`
	TimeoutSec int    `json:"timeout_seconds"`
}

// ScriptAtomicRelation describes the atomic task relation.
// Relations defines the order of atomic tasks execution.
type ScriptAtomicRelation struct {
	AtomicTaskID int   `json:"atomic_task_id"`
	Dependencies []int `json:"atomic_task_id_idx"`
}

// ReqExecuteScript describes the request data of execute_script.
type ReqExecuteScript struct {
	Endpoints []*EndpointWithAuth     `json:"agents"`
	Scripts   []*ScriptDetail         `json:"scripts"`
	Atomics   []*ScriptAtomicTask     `json:"atomic_tasks"`
	Relations []*ScriptAtomicRelation `json:"atomic_tasks_relations"`
}

// ScriptResult describes the script result.
type ScriptResult struct {
	Endpoint

	AtomicTaskID int    `json:"atomic_task_id"`
	Status       int    `json:"status"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_msg"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	ExitCode     int    `json:"script_exit_code"`
	Tag          string `json:"tag"`
	ScreenLog    string `json:"screen"`
}

// RespExecuteScript describes the response data of execute_script.
type RespExecuteScript struct {
	RespCommon

	Data struct {
		Result struct {
			TaskID string `json:"task_id"`
		} `json:"result"`
	} `json:"data"`
}

// ScriptAtomicTaskCondition describes the query condition of script atomic task.
type ScriptAtomicTaskCondition struct {
	ID     int  `json:"atomic_task_id"`
	Offset uint `json:"offset"`
	Limit  uint `json:"limit"`
}

// ScriptEndpointCondition describes the query endpoint condition of script atomic task.
type ScriptEndpointCondition struct {
	Endpoint

	Atomics []*ScriptAtomicTaskCondition `json:"atomic_tasks"`
}

// ReqQueryScriptExecutionResult describes the request data of query_script_execution_result.
type ReqQueryScriptExecutionResult struct {
	TaskID string `json:"task_id"`

	AgentTasks []*ScriptEndpointCondition `json:"agent_tasks"`
}

// RespQueryScriptExecutionResult describes the response data of query_script_execution_result.
type RespQueryScriptExecutionResult struct {
	RespCommon

	Data struct {
		Result []*ScriptResult `json:"result"`
	} `json:"data"`
}

// ReqTerminateScriptExecution describes the request data of terminate_script_execution.
type ReqTerminateScriptExecution struct {
	TaskID    string      `json:"task_id"`
	Endpoints []*Endpoint `json:"agents"`
}

// RespTerminateScriptExecution describes the response data of terminate_script_execution.
type RespTerminateScriptExecution struct {
	RespCommon

	Data struct {
		Result struct {
			TaskID string `json:"task_id"`
		} `json:"result"`
	} `json:"data"`
}

// TransferSource describes the transfer source.
type TransferSource struct {
	FileName  string           `json:"file_name"`
	StoredDir string           `json:"store_dir"`
	MD5       string           `json:"md5"`
	Endpoint  EndpointWithAuth `json:"agent"`
}

// TransferTarget describes the transfer target.
type TransferTarget struct {
	FileName   string              `json:"file_name"`
	StoredDir  string              `json:"store_dir"`
	Owner      string              `json:"owner"`
	Permission int                 `json:"permission"`
	Endpoints  []*EndpointWithAuth `json:"agents"`
}

// TransferDetail describes the transfer detail.
type TransferDetail struct {
	Source *TransferSource `json:"source"`
	Target *TransferTarget `json:"target"`
}

// ReqTransferFile describes the request data of transfer_file.
type ReqTransferFile struct {
	TimeoutSec    uint              `json:"timeout_seconds"`
	AutoMkdir     bool              `json:"auto_mkdir"`
	UploadSpeed   uint              `json:"upload_speed"`
	DownloadSpeed uint              `json:"download_speed"`
	Tasks         []*TransferDetail `json:"tasks"`
}

// RespTransferFile describes the response data of transfer_file.
type RespTransferFile struct {
	RespCommon

	Data struct {
		Result struct {
			TaskID string `json:"task_id"`
		} `json:"result"`
	} `json:"data"`
}

// FileResultContent describes the file result content.
type FileResultContent struct {
	DestinationAgentID     string `json:"dest_agent_id"`
	DestinationContainerID string `json:"dest_container_id"`
	DestFileDir            string `json:"dest_file_dir"`
	DestFileName           string `json:"dest_file_name"`
	Mode                   int    `json:"mode"`
	Progress               uint   `json:"progress"`
	SourceAgentID          string `json:"source_agent_id"`
	SourceContainerID      string `json:"source_container_id"`
	SourceFileDir          string `json:"source_file_dir"`
	SourceFileName         string `json:"source_file_name"`
	Speed                  uint64 `json:"speed"`
	Status                 int    `json:"status"`
	StatusInfo             string `json:"status_info"`
	Type                   string `json:"type"`
	StartTime              int64  `json:"start_time"`
	EndTime                int64  `json:"end_time"`
	Size                   uint64 `json:"size"`
}

// FileResult describes the file result.
type FileResult struct {
	ErrorCode    int               `json:"error_code"`
	ErrorMessage string            `json:"error_msg"`
	Content      FileResultContent `json:"content"`
}

// ReqQueryFileTransmissionResult describes the request data of query_file_transmission_result.
type ReqQueryFileTransmissionResult struct {
	TaskID    string      `json:"task_id"`
	Endpoints []*Endpoint `json:"agents"`
}

// RespQueryFileTransmissionResult describes the response data of query_file_transmission_result.
type RespQueryFileTransmissionResult struct {
	RespCommon

	Data struct {
		Result []*FileResult `json:"result"`
	} `json:"data"`
}

// ReqTerminateFileTransmission describes the request data of terminate_file_transmission.
type ReqTerminateFileTransmission struct {
	TaskID    string      `json:"task_id"`
	Endpoints []*Endpoint `json:"agents"`
}

// RespTerminateFileTransmission describes the response data of terminate_file_transmission.
type RespTerminateFileTransmission struct {
	RespCommon

	Data struct {
		Result struct {
			TaskID string `json:"task_id"`
		} `json:"result"`
	} `json:"data"`
}
