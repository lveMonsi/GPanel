package models

// SessionHistory 终端会话历史记录
type SessionHistory struct {
	BaseModel
	HostID    uint   `json:"hostID" gorm:"index"` // 关联的主机ID（0表示本地终端）
	HostAddr  string `json:"hostAddr" gorm:"type:varchar(256)"` // 主机地址（用于记录）
	UserName  string `json:"userName" gorm:"type:varchar(64)"` // 连接用户名
	StartTime int64  `json:"startTime"` // 开始时间戳
	EndTime   int64  `json:"endTime"`   // 结束时间戳
	Duration  int    `json:"duration"`  // 会话时长（秒）
	Commands  string `json:"commands" gorm:"type:text"` // 执行的命令（JSON格式）
}