package dto

// OperationLogCreate 创建操作日志
type OperationLogCreate struct {
	Username string `json:"username" binding:"required"`
	IP       string `json:"ip"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
	Detail   string `json:"detail"`
	Status   string `json:"status" binding:"required"`
}

// OperationLogSearch 搜索操作日志
type OperationLogSearch struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	Username  string `json:"username"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	Status    string `json:"status"`
	Keyword   string `json:"keyword"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// OperationLogClean 清理操作日志
type OperationLogClean struct {
	RetainDays int `json:"retainDays" binding:"required,min=1"`
}

// OperationLogStats 操作日志统计
type OperationLogStats struct {
	Total        int64            `json:"total"`
	TodayCount   int64            `json:"todayCount"`
	ResourceStat map[string]int64 `json:"resourceStat"`
}

// SystemLogSearch 搜索系统日志
type SystemLogSearch struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword"`
	LogFile  string `json:"logFile"`
}

// SystemLogInfo 系统日志文件信息
type SystemLogInfo struct {
	Files []SystemLogFile `json:"files"`
}

// SystemLogFile 系统日志文件
type SystemLogFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}
