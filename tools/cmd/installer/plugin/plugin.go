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

// Package plugin provides plugin related commands.
package plugin

import (
	"fmt"
	"runtime"

	pluginFlag "github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/plugin/step"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/utils"
	"github.com/spf13/cobra"
)

// NewPluginCommand creates a new plugin sub command.
// nolint: lll
func NewPluginCommand() *cobra.Command {
	pluginCommand := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin command",
		Long:  "Plugin command",
		// PersistentPreRunE is inherited by every plugin leaf command (none of them defines
		// its own PersistentPreRunE) and runs before any HTTP request, so dropping the
		// proxy env vars here makes every http.Client skip the HTTP proxy.
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			skipHTTPProxy, err := cmd.Flags().GetBool(pluginFlag.SkipHTTPProxy)
			if err != nil {
				return err
			}

			if !skipHTTPProxy {
				return nil
			}

			fmt.Printf("flag --%s is set, all http requests will skip the proxy\n", pluginFlag.SkipHTTPProxy)

			return utils.UnsetProxyEnv()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	pluginCommand.AddCommand(step.NewStepCommand())
	pluginCommand.AddCommand(NewFullInstall())
	pluginCommand.AddCommand(NewFullUninstall())
	pluginCommand.AddCommand(NewFullUpgrade())
	pluginCommand.AddCommand(NewFullDebug())

	/*
	 * persistent required flags.
	 */
	pluginCommand.PersistentFlags().StringP(pluginFlag.DeployEnv, pluginFlag.DeployEnvS, "", "the deploy environment which to operate at")
	_ = pluginCommand.MarkFlagRequired(pluginFlag.DeployEnv)

	/*
	 * persistent optional flags.
	 */
	pluginCommand.PersistentFlags().String(pluginFlag.PluginName, "", "plugin name")
	pluginCommand.PersistentFlags().String(pluginFlag.PluginGroup, "", "plugin group, this will be used to build the plugin directory")
	pluginCommand.PersistentFlags().String(pluginFlag.PluginPkgName, "", "plugin package name, this will be used to download the plugin package")
	pluginCommand.PersistentFlags().String(pluginFlag.BaseDeployDir, defaultBaseDeployDir(), "base deployed directory of this node, the deploy dir will be created under this directory with deploy-env")
	pluginCommand.PersistentFlags().String(pluginFlag.BaseWorkDir, defaultBaseWorkDir(), "base work directory of this node, the work dir will be created under this directory with deploy-env")
	pluginCommand.PersistentFlags().Bool(pluginFlag.SkipHTTPProxy, false, "skip the http proxy for all http requests in the installer")

	return pluginCommand
}

func defaultBaseDeployDir() string {
	if runtime.GOOS == "windows" {
		return `C:\`
	}

	return "/usr/local/"
}

func defaultBaseWorkDir() string {
	if runtime.GOOS == "windows" {
		return `C:\tmp\bknm\`
	}

	return "/tmp/bknm/"
}
