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

// Package blog provides a logging wrapper for glog.
package blog

import (
	"log"
	"sync"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/blog/glog"
)

// LogConfig is the configuration for initializing logs.
type LogConfig glog.LogConfig

func NewLogConfig() LogConfig {
	return LogConfig{
		LogDir:       "./logs",
		LogMaxSizeMB: 200,
		LogMaxNum:    10,

		ToStdErr:        true,
		AlsoToStdErr:    true,
		Level:           "info",
		StdErrThreshold: "3",
		VModule:         "",
		TraceLocation:   "",
	}
}

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

var once sync.Once

// InitLogs initializes logs the way we want for blog.
func InitLogs(logConfig LogConfig) {
	glog.InitLogs(glog.LogConfig(logConfig))

	once.Do(func() {
		log.SetOutput(GlogWriter{})
		log.SetFlags(0)
		// The default glog flush interval is 30 seconds, which is frighteningly long.
		go func() {
			d := time.Duration(5 * time.Second) // nolint
			tick := time.Tick(d)

			for range tick {
				glog.Flush()
			}
		}()
	})
}

// CloseLogs flush the rest logs in buffer.
func CloseLogs() {
	glog.Flush()
}

var (
	// Debug prints logs like fmt.Print in debug level.
	Debug = glog.Debug
	// Debugf prints logs like fmt.Printf in debug level.
	Debugf = glog.Debugf
	// Debugw prints logs with key(values) in debug level.
	Debugw = glog.Debugw

	// Info prints logs like fmt.Print in info level.
	Info = glog.Info
	// Infof prints logs like fmt.Printf in info level.
	Infof = glog.Infof
	// Infow prints logs with key(values) in info level.
	Infow = glog.Infow

	// Warn prints logs like fmt.Print in warn level.
	Warn = glog.Warning
	// Warnf prints logs like fmt.Printf in warn level.
	Warnf = glog.Warningf
	// Warnw prints logs with key(values) in warn level.
	Warnw = glog.Warningw

	// Error prints	logs like fmt.Print in error level.
	Error = glog.Error
	// Errorf prints logs like fmt.Printf in error level.
	Errorf = glog.Errorf
	// Errorw prints logs with key(values) in error level.
	Errorw = glog.Errorw
)

// SetLevel set the logging level.
func SetLevel(level string) {
	glog.SetLevel(level)
}
