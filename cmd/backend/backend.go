/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main of backend.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"git.woa.com/bk-gse/bk-nodeman/internal/service/backend"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

const (
	// ModeDebug debug mode
	ModeDebug = "debug"
	// ModeRelease release mode
	ModeRelease = "release"
)

var (
	// configPath of backend service
	configPath string

	// mode of backend service
	mode string

	serverCmd = &cobra.Command{
		Use:   "bk_nodeman_backend",
		Short: "bk-nodeman backend server",
		Long:  "bk-nodeman backend server",
		Run: func(_ *cobra.Command, _ []string) {
			config := &config.BackendService{}
			if err := config.LoadFromFile(configPath); err != nil {
				fmt.Printf("failed to load config file(%s): %v\n", configPath, err)
				os.Exit(1)
			}

			if err := config.Validate(); err != nil {
				fmt.Printf("failed to validate config: %v\n", err)
				os.Exit(1)
			}

			switch mode {
			case ModeDebug:
				gin.SetMode(gin.DebugMode)
			case ModeRelease:
				gin.SetMode(gin.ReleaseMode)
				gin.DebugPrintFunc = func(format string, args ...interface{}) {
					fmt.Fprintf(blog.GlogWriter{}, format, args...)
				}
			default:
				fmt.Printf("invalid mode: %s\n", mode)
				os.Exit(1)
			}

			if err := backend.NewService(config).Start(context.Background()); err != nil {
				fmt.Printf("failed to start service: %v\n", err)
				os.Exit(1)
			}

			// listening signal
			// TODO: handle SIGTERM and do graceful shutdown.
			signalC := make(chan os.Signal, 1)
			signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)
			receivedSignal := <-signalC

			fmt.Printf("received signal(%s), going to exit server\n", receivedSignal.String())
		},
	}
)

// backend service entrypoint.
func main() {
	serverCmd.PersistentFlags().StringVarP(
		&configPath, "file", "f", "", "path of service config file",
	)

	err := serverCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		fmt.Printf("failed to mark flag required, err: %v\n", err)
		os.Exit(1)
	}

	serverCmd.PersistentFlags().StringVarP(
		&mode, "mode", "m", ModeRelease, fmt.Sprintf("set mode of backend service, support %s",
			[]string{ModeDebug, ModeRelease}),
	)

	err = serverCmd.Execute()
	if err != nil {
		fmt.Printf("failed to execute cmd, err: %v\n", err)
		os.Exit(1)
	}
}
