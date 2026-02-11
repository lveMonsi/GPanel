package dto

// SessionHistoryCreate 创建会话历史
type SessionHistoryCreate struct {
	HostID   uint   `json:"hostID"`
	HostAddr string `json:"hostAddr"`
	UserName string `json:"userName"`
}

// SessionHistoryUpdate 更新会话历史
type SessionHistoryUpdate struct {
	ID       uint   `json:"id"`
	EndTime  int64  `json:"endTime"`
	Duration int    `json:"duration"`
	Commands string `json:"commands"`
}

// SessionHistorySearch 搜索会话历史
type SessionHistorySearch struct {
	PageInfo
	HostID   uint   `json:"hostID"`   // 主机ID筛选
	HostAddr string `json:"hostAddr"` // 主机地址筛选
	UserName string `json:"userName"` // 用户名筛选
	StartDate int64 `json:"startDate"` // 开始日期筛选
	EndDate   int64 `json:"endDate"`   // 结束日期筛选
}

// SessionHistoryInfo 会话历史信息
type SessionHistoryInfo struct {
	ID        uint   `json:"id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	HostID    uint   `json:"hostID"`
	HostAddr  string `json:"hostAddr"`
	UserName  string `json:"userName"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Duration  int    `json:"duration"`
	Commands  string `json:"commands"`
}