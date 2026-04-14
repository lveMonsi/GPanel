package models

import "time"

// OperationLog 操作日志模型
type OperationLog struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Username  string    `json:"username" gorm:"type:varchar(64);index"`
	IP        string    `json:"ip" gorm:"type:varchar(64)"`
	Resource  string    `json:"resource" gorm:"type:varchar(64);index"`
	Action    string    `json:"action" gorm:"type:varchar(32);index"`
	Detail    string    `json:"detail" gorm:"type:text"`
	Status    string    `json:"status" gorm:"type:varchar(16);index"`
}
