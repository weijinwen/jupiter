// Copyright 2026 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package governor

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/weijinwen/jupiter/pkg"
)

// SystemStats 系统运行时监控信息。
type SystemStats struct {
	// 系统基础信息
	Hostname  string `json:"hostname"`
	GoVersion string `json:"goVersion"`
	StartTime string `json:"startTime"`
	Uptime    string `json:"uptime"`

	// CPU
	CPUCores   int     `json:"cpuCores"`
	GOMAXPROCS int     `json:"gomaxprocs"`
	CPUUsage   float64 `json:"cpuUsage"`

	// 系统内存
	TotalMemory     uint64  `json:"totalMemory"`
	UsedMemory      uint64  `json:"usedMemory"`
	AvailableMemory uint64  `json:"availableMemory"`
	MemoryUsage     float64 `json:"memoryUsage"`

	// Go 运行时内存
	HeapAlloc    uint64 `json:"heapAlloc"`
	HeapSys      uint64 `json:"heapSys"`
	HeapIdle     uint64 `json:"heapIdle"`
	HeapInuse    uint64 `json:"heapInuse"`
	HeapReleased uint64 `json:"heapReleased"`
	StackInuse   uint64 `json:"stackInuse"`
	StackSys     uint64 `json:"stackSys"`
	TotalAlloc   uint64 `json:"totalAlloc"`
	Sys          uint64 `json:"sys"`
	NumGC        uint32 `json:"numGC"`

	// Goroutine
	NumGoroutine int `json:"numGoroutine"`
}

// startTime 是应用启动时间，由 collectStats 首次调用时初始化。
var startTime time.Time

func init() {
	startTime = time.Now()
}

// collectStats 采集系统运行时监控信息。
func collectStats() *SystemStats {
	stats := &SystemStats{
		Hostname:     pkg.HostName(),
		GoVersion:    pkg.GoVersion(),
		StartTime:    pkg.StartTime(),
		Uptime:       time.Since(startTime).Truncate(time.Second).String(),
		CPUCores:     runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		NumGoroutine: runtime.NumGoroutine(),
	}

	// CPU 使用率（采样 1 秒，取所有核心的平均值）
	if percents, err := cpu.Percent(time.Second, false); err == nil && len(percents) > 0 {
		stats.CPUUsage = percents[0]
	}

	// 系统内存
	if v, err := mem.VirtualMemory(); err == nil {
		stats.TotalMemory = v.Total
		stats.UsedMemory = v.Used
		stats.AvailableMemory = v.Available
		stats.MemoryUsage = v.UsedPercent
	}

	// Go 运行时内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats.HeapAlloc = m.HeapAlloc
	stats.HeapSys = m.HeapSys
	stats.HeapIdle = m.HeapIdle
	stats.HeapInuse = m.HeapInuse
	stats.HeapReleased = m.HeapReleased
	stats.StackInuse = m.StackInuse
	stats.StackSys = m.StackSys
	stats.TotalAlloc = m.TotalAlloc
	stats.Sys = m.Sys
	stats.NumGC = m.NumGC

	return stats
}
