package models

import "time"

type MonitorData struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`

	// CPU
	CPUPercent float64 `json:"cpuPercent"`

	// 内存
	MemTotal   uint64  `json:"memTotal"`
	MemUsed    uint64  `json:"memUsed"`
	MemPercent float64 `json:"memPercent"`

	// 负载
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`

	// 磁盘IO
	DiskReadBytes  uint64 `json:"diskReadBytes"`
	DiskWriteBytes uint64 `json:"diskWriteBytes"`

	// 网络IO
	NetRecvBytes uint64 `json:"netRecvBytes"`
	NetSentBytes uint64 `json:"netSentBytes"`
}

type MonitorSetting struct {
	ID              uint `gorm:"primaryKey" json:"id"`
	Enabled         bool `json:"enabled"`
	RetentionDays   int  `json:"retentionDays"`
	CollectInterval int  `json:"collectInterval"` // 秒
}
