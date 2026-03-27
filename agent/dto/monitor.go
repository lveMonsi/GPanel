package dto

import "time"

type MonitorQueryReq struct {
	Param     string    `json:"param" binding:"required,oneof=load cpu memory io network all"`
	IO        string    `json:"io"`
	Network   string    `json:"network"`
	StartTime time.Time `json:"startTime" binding:"required"`
	EndTime   time.Time `json:"endTime" binding:"required"`
}

type Process struct {
	Name    string  `json:"name"`
	Pid     int32   `json:"pid"`
	Percent float64 `json:"percent"`
	Memory  uint64  `json:"memory"`
	User    string  `json:"user"`
}

type MonitorBaseData struct {
	Date      time.Time `json:"date"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Load1     float64   `json:"load1"`
	Load5     float64   `json:"load5"`
	Load15    float64   `json:"load15"`
	LoadUsage float64   `json:"loadUsage"`
	TopCPU    []Process `json:"topCPU,omitempty"`
	TopMem    []Process `json:"topMem,omitempty"`
}

type MonitorIOData struct {
	Date  time.Time `json:"date"`
	Read  uint64    `json:"read"`
	Write uint64    `json:"write"`
	Count uint64    `json:"count"`
	Time  uint64    `json:"time"`
}

type MonitorNetworkData struct {
	Date time.Time `json:"date"`
	Up   float64   `json:"up"`
	Down float64   `json:"down"`
}

type MonitorSettingResp struct {
	Enabled         bool   `json:"enabled"`
	RetentionDays   int    `json:"retentionDays"`
	CollectInterval int    `json:"collectInterval"`
	DefaultNetwork  string `json:"defaultNetwork"`
	DefaultIO       string `json:"defaultIO"`
}

type MonitorSettingReq struct {
	Enabled         *bool   `json:"enabled"`
	RetentionDays   *int    `json:"retentionDays"`
	CollectInterval *int    `json:"collectInterval"`
	DefaultNetwork  *string `json:"defaultNetwork"`
	DefaultIO       *string `json:"defaultIO"`
}

