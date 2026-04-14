package dto

// CronjobCreate 创建计划任务
type CronjobCreate struct {
	Name           string `json:"name" binding:"required"`
	Type           string `json:"type" binding:"required"`
	Spec           string `json:"spec" binding:"required"`
	SpecCustom     bool   `json:"specCustom"`
	Script         string `json:"script"`
	URL            string `json:"url"`
	SourceDir      string `json:"sourceDir"`
	ExclusionRules string `json:"exclusionRules"`
	RetainCopies   int    `json:"retainCopies"`
	RetryCount     int    `json:"retryCount"`
	Timeout        int    `json:"timeout"`
	IgnoreErr      bool   `json:"ignoreErr"`
}

// CronjobUpdate 更新计划任务
type CronjobUpdate struct {
	ID             uint   `json:"id" binding:"required"`
	Name           string `json:"name"`
	Spec           string `json:"spec"`
	SpecCustom     bool   `json:"specCustom"`
	Script         string `json:"script"`
	URL            string `json:"url"`
	SourceDir      string `json:"sourceDir"`
	ExclusionRules string `json:"exclusionRules"`
	RetainCopies   int    `json:"retainCopies"`
	RetryCount     int    `json:"retryCount"`
	Timeout        int    `json:"timeout"`
	IgnoreErr      bool   `json:"ignoreErr"`
}

// CronjobDelete 批量删除
type CronjobDelete struct {
	IDs []uint `json:"ids" binding:"required"`
}

// CronjobSearch 搜索
type CronjobSearch struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword"`
	Type     string `json:"type"`
	Status   string `json:"status"`
}

// CronjobToggle 启用/禁用
type CronjobToggle struct {
	ID     uint   `json:"id" binding:"required"`
	Status string `json:"status" binding:"required"`
}

// CronjobHandle 手动执行
type CronjobHandle struct {
	ID uint `json:"id" binding:"required"`
}

// CronjobStop 停止执行
type CronjobStop struct {
	ID uint `json:"id" binding:"required"`
}

// RecordSearch 搜索执行记录
type RecordSearch struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	CronjobID uint   `json:"cronjobId" binding:"required"`
	Status    string `json:"status"`
}

// RecordClean 清空记录
type RecordClean struct {
	CronjobID uint `json:"cronjobId" binding:"required"`
}

// NextTimesReq 获取下次执行时间
type NextTimesReq struct {
	Spec string `json:"spec" binding:"required"`
}
