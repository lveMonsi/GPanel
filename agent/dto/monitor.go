package dto

import "time"

type MonitorDataResp struct {
	Timestamp      time.Time `json:"timestamp"`
	CPUPercent     float64   `json:"cpuPercent"`
	MemTotal       uint64    `json:"memTotal"`
	MemUsed        uint64    `json:"memUsed"`
	MemPercent     float64   `json:"memPercent"`
	Load1          float64   `json:"load1"`
	Load5          float64   `json:"load5"`
	Load15         float64   `json:"load15"`
	DiskReadBytes  uint64    `json:"diskReadBytes"`
	DiskWriteBytes uint64    `json:"diskWriteBytes"`
	NetRecvBytes   uint64    `json:"netRecvBytes"`
	NetSentBytes   uint64    `json:"netSentBytes"`
}

type MonitorQueryReq struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type MonitorSettingResp struct {
	Enabled         bool `json:"enabled"`
	RetentionDays   int  `json:"retentionDays"`
	CollectInterval int  `json:"collectInterval"`
}

type MonitorSettingReq struct {
	Enabled         *bool `json:"enabled"`
	RetentionDays   *int  `json:"retentionDays"`
	CollectInterval *int  `json:"collectInterval"`
}
