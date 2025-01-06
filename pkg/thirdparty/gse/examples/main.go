/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main for gse example.
package main

import (
	"fmt"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/apigw"
	"git.woa.com/bk-gse/bk-nodeman/pkg/thirdparty/gse"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
)

const (
	agentID1 string = ""
	agentID2 string = ""
	agentID3 string = ""
	taskID   string = ""

	executionTimeout          = 10 * time.Minute
	transmissionTimeout       = 10 * time.Minute
	transmissionSpeedMBPerSec = 10
)

func listAgentInfo(gseHandler gse.Handler) {
	result, err := gseHandler.ListAgentInfo([]string{
		agentID1,
		agentID2,
		agentID3,
	})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("agent info")
	for _, info := range result {
		fmt.Println("agent_id:", info.AgentID)
		fmt.Println("cloud_id:", info.CloudID)
		fmt.Println("version:", info.Version)
		fmt.Println("report_time:", info.ReportTime)
		fmt.Println("status_code:", info.StatusCode)
		fmt.Println("status:", info.Status)
		fmt.Println("last_status_code:", info.LastStatusCode)
		fmt.Println("last_status:", info.LastStatus)
		fmt.Println("host_ip:", info.HostIP)
		fmt.Println("parent_ip:", info.ParentIP)
		fmt.Println("parent_port:", info.ParentPort)
		fmt.Println()
	}
}

func listAgentState(gseHandler gse.Handler) {
	result, err := gseHandler.ListAgentState([]string{
		agentID1,
		agentID2,
		agentID3,
	})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println("agent state")
	for _, state := range result {
		fmt.Println("agent_id:", state.AgentID)
		fmt.Println("cloud_id:", state.CloudID)
		fmt.Println("version:", state.Version)
		fmt.Println("report_time:", state.ReportTime)
		fmt.Println("status_code:", state.StatusCode)
		fmt.Println("last_status_code:", state.LastStatusCode)
		fmt.Println()
	}
}

func executeScript(gseHandler gse.Handler) {
	result, err := gseHandler.ExecuteScript([]*types.EndpointWithAuth{
		{
			Endpoint: types.Endpoint{
				AgentID: agentID1,
			},

			User: "root",
		},
		{
			Endpoint: types.Endpoint{
				AgentID: agentID2,
			},

			User: "root",
		},
		{
			Endpoint: types.Endpoint{
				AgentID: agentID3,
			},

			User: "root",
		},
	}, "echo 1", executionTimeout)
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(result)
}

func queryScriptExecutionResult(gseHandler gse.Handler) {
	result, err := gseHandler.QueryScriptExecutionResult(taskID, []*types.EndpointWithRestrict{
		{
			Endpoint: types.Endpoint{
				AgentID: agentID1,
			},
			Offset: 0,
			Limit:  0,
		},
	})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(result)
}

func terminateScriptExecution(gseHandler gse.Handler) {
	result, err := gseHandler.TerminateScriptExecution(taskID, []*types.Endpoint{
		{
			AgentID: agentID1,
		},
		{
			AgentID: agentID2,
		},
		{
			AgentID: agentID3,
		},
	})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(result)
}

func transferFile(gseHandler gse.Handler) {
	result, err := gseHandler.TransferFile(&types.TransferOptions{
		Timeout:               transmissionTimeout,
		AutoMkdir:             true,
		UploadSpeedMBPerSec:   transmissionSpeedMBPerSec,
		DownloadSpeedMBPerSec: transmissionSpeedMBPerSec,
	}, []*types.TransferDetail{
		{
			Source: types.TransferSource{
				FileName:  "testfile",
				StoredDir: "/tmp/",
				Endpoint: types.EndpointWithAuth{
					Endpoint: types.Endpoint{
						AgentID: agentID1,
					},
					User: "root",
				},
			},
			Target: types.TransferTarget{
				StoredDir: "/tmp/testdir",
				Endpoints: []*types.EndpointWithAuth{
					{
						Endpoint: types.Endpoint{
							AgentID: agentID2,
						},
						User: "root",
					},
					{
						Endpoint: types.Endpoint{
							AgentID: agentID3,
						},
						User: "root",
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(result)
}

func queryFileTransmissionResult(gseHandler gse.Handler) {
	result, err := gseHandler.QueryFileTransmissionResult(taskID, []*types.Endpoint{
		{
			AgentID: agentID1,
		},
		{
			AgentID: agentID2,
		},
		{
			AgentID: agentID3,
		},
	})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(result)
}

func terminateFileTransmission(gseHandler gse.Handler) {
	result, err := gseHandler.TerminateFileTransmission(taskID, []*types.Endpoint{
		{
			AgentID: agentID1,
		},
		{
			AgentID: agentID2,
		},
		{
			AgentID: agentID3,
		}})
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Println(result)
}

func main() {
	apigwCli := apigw.NewClient(&apigw.Config{
		BKAppCode:        "",
		BKAppSecret:      "",
		PlatformUsername: "admin",
		APIGWDomain:      "",
	})
	gseHandler := gse.NewHandler(&gse.Config{Environment: "prod"}, apigwCli)

	listAgentInfo(gseHandler)
	listAgentState(gseHandler)
	executeScript(gseHandler)
	terminateScriptExecution(gseHandler)
	queryScriptExecutionResult(gseHandler)
	transferFile(gseHandler)
	terminateFileTransmission(gseHandler)
	queryFileTransmissionResult(gseHandler)
}
