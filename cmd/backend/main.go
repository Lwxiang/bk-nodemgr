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

	"git.woa.com/bk-gse/bk-nodeman/internal/backend/service"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	machinerylog "github.com/RichardKnop/machinery/v2/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func init() {
	gin.DebugPrintFunc = func(format string, args ...interface{}) {
		fmt.Fprintf(blog.WriterDebug{}, format, args...)
	}

	machinerylog.Set(logger{})
}

type logger struct{}

func (l logger) Print(args ...interface{}) {
	blog.Info(args...)
}

func (l logger) Printf(s string, args ...interface{}) {
	blog.Infof(s, args...)
}

func (l logger) Println(args ...interface{}) {
	blog.Info(args...)
}

func (l logger) Fatal(args ...interface{}) {
	blog.Error(args...)
}

func (l logger) Fatalf(s string, args ...interface{}) {
	blog.Errorf(s, args...)
}

func (l logger) Fatalln(args ...interface{}) {
	blog.Error(args...)
}

func (l logger) Panic(args ...interface{}) {
	blog.Error(args...)
}

func (l logger) Panicf(s string, args ...interface{}) {
	blog.Errorf(s, args...)
}

func (l logger) Panicln(args ...interface{}) {
	blog.Error(args...)
}

// backend service entrypoint.
func main() {
	// configPath of backend service.
	var configPath string

	serverCmd := &cobra.Command{
		Use:   "bk_nodeman_backend",
		Short: "bk-nodeman backend server",
		Long:  "bk-nodeman backend server",
		Run: func(_ *cobra.Command, _ []string) {
			conf := config.NewBackendService()
			if err := conf.LoadFromFile(configPath); err != nil {
				fmt.Printf("failed to load config file(%s): %v\n", configPath, err)
				os.Exit(1)
			}

			if err := conf.Validate(); err != nil {
				fmt.Printf("failed to validate config: %v\n", err)
				os.Exit(1)
			}

			switch conf.RunMode {
			case config.RunModeDebug:
				gin.SetMode(gin.DebugMode)
			case config.RunModeRelease:
				gin.SetMode(gin.ReleaseMode)
			default:
				fmt.Printf("invalid mode: %s\n", conf.RunMode)
				os.Exit(1)
			}

			service, err := service.NewService(conf)
			if err != nil {
				fmt.Printf("failed to create service: %v\n", err)
				os.Exit(1)
			}

			if err := service.Start(context.Background()); err != nil {
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

	serverCmd.PersistentFlags().StringVarP(
		&configPath, "file", "f", "", "path of service config file",
	)

	err := serverCmd.MarkPersistentFlagRequired("file")
	if err != nil {
		fmt.Printf("failed to mark flag required, err: %v\n", err)
		os.Exit(1)
	}

	err = serverCmd.Execute()
	if err != nil {
		fmt.Printf("failed to execute cmd, err: %v\n", err)
		os.Exit(1)
	}
}
