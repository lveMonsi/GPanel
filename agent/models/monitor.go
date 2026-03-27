package models

import "time"

type MonitorBase struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
	CPU        float64   `json:"cpu"`
	Memory     float64   `json:"memory"`
	Load1      float64   `json:"load1"`
	Load5      float64   `json:"load5"`
	Load15     float64   `json:"load15"`
	LoadUsage  float64   `json:"loadUsage"`
	TopCPU     string    `gorm:"type:text" json:"topCPU"`
	TopMem     string    `gorm:"type:text" json:"topMem"`
}

type MonitorIO struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
	Name      string    `gorm:"index" json:"name"`
	Read      uint64    `json:"read"`
	Write     uint64    `json:"write"`
	Count     uint64    `json:"count"`
	Time      uint64    `json:"time"`
}

type MonitorNetwork struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
	Name      string    `gorm:"index" json:"name"`
	Up        float64   `json:"up"`
	Down      float64   `json:"down"`
}

type MonitorSetting struct {
	ID              uint `gorm:"primaryKey" json:"id"`
	Enabled         bool `json:"enabled"`
	RetentionDays   int  `json:"retentionDays"`
	CollectInterval int  `json:"collectInterval"`
	DefaultNetwork  string `json:"defaultNetwork"`
	DefaultIO       string `json:"defaultIO"`
}
