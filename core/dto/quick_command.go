package dto

// QuickCommandCreate 创建快速命令
type QuickCommandCreate struct {
	Name        string `json:"name" binding:"required"`
	Command     string `json:"command" binding:"required"`
	Description string `json:"description"`
	GroupID     uint   `json:"groupId"`
	Sort        int    `json:"sort"`
}

// QuickCommandUpdate 更新快速命令
type QuickCommandUpdate struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
	GroupID     uint   `json:"groupId"`
	Sort        int    `json:"sort"`
}

// QuickCommandDelete 删除快速命令
type QuickCommandDelete struct {
	IDs []uint `json:"ids" binding:"required"`
}

// QuickCommandSearch 搜索快速命令
type QuickCommandSearch struct {
	PageInfo
	Keyword string `json:"keyword"`
	GroupID uint   `json:"groupId"`
}

// QuickCommandPageResult 快速命令分页结果
type QuickCommandPageResult struct {
	Items []QuickCommandItem `json:"items"`
	Total int64              `json:"total"`
}

// QuickCommandItem 快速命令项
type QuickCommandItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
	GroupID     uint   `json:"groupId"`
	Sort        int    `json:"sort"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}