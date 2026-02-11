package models

// QuickCommand 快速命令模型
type QuickCommand struct {
	BaseModel
	Name        string `json:"name" gorm:"type:varchar(100);not null;index:idx_name"`
	Command     string `json:"command" gorm:"type:text;not null"`
	Description string `json:"description" gorm:"type:varchar(500)"`
	GroupID     uint   `json:"groupId" gorm:"default:0;index:idx_group"`
	Sort        int    `json:"sort" gorm:"default:0;index:idx_sort"`
}