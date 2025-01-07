/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package glog

import (
	"strings"
	"sync"
)

// SetLevel set the logging level.
func SetLevel(level string) {
	var value int32

	switch strings.ToUpper(level) {
	case "DEBUG":
		value = int32(debugLog)
	case "INFO":
		value = int32(infoLog)
	case "WARN":
		value = int32(warningLog)
	case "ERROR":
		value = int32(errorLog)
	default:
		value = int32(infoLog)
	}

	_ = logging.verbosity.Set(value)
}

var once sync.Once

// LogConfig defines the glog config.
type LogConfig struct {
	LogDir       string
	LogMaxSizeMB int
	LogMaxNum    int

	ToStdErr        bool
	AlsoToStdErr    bool
	Level           string
	StdErrThreshold string
	VModule         string
	TraceLocation   string
}

// InitLogs init glog from params.
func InitLogs(config LogConfig) {
	once.Do(func() {
		logging.toStderr = config.ToStdErr
		logging.alsoToStderr = config.AlsoToStdErr
		_ = logging.stderrThreshold.Set(config.StdErrThreshold)
		_ = logging.vmodule.Set(config.VModule)
		_ = logging.traceLocation.Set(config.TraceLocation)

		SetLevel(config.Level)

		logMaxNum = config.LogMaxNum
		logMaxSize = uint64(config.LogMaxSizeMB) * 1024 * 1024
		logDir = config.LogDir
	})
}
