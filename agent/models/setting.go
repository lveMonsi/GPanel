package models

import "time"

type BaseModel struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Setting 系统设置模型
type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"type:varchar(256);not null;uniqueIndex"`
	Value string `json:"value" gorm:"type:text"`
	About string `json:"about" gorm:"type:text"`
}