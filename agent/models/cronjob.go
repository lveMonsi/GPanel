package models

import "time"

// Cronjob 计划任务模型
type Cronjob struct {
	ID             uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Name           string    `json:"name" gorm:"type:varchar(100);not null"`
	Type           string    `json:"type" gorm:"type:varchar(20);not null;index:idx_type"`
	Spec           string    `json:"spec" gorm:"type:varchar(200);not null"`
	SpecCustom     bool      `json:"specCustom" gorm:"default:false"`
	Script         string    `json:"script" gorm:"type:text"`
	URL            string    `json:"url" gorm:"type:varchar(500)"`
	SourceDir      string    `json:"sourceDir" gorm:"type:varchar(500)"`
	ExclusionRules string    `json:"exclusionRules" gorm:"type:text"`
	RetainCopies   int       `json:"retainCopies" gorm:"default:5"`
	RetryCount     int       `json:"retryCount" gorm:"default:0"`
	Timeout        int       `json:"timeout" gorm:"default:0"`
	IgnoreErr      bool      `json:"ignoreErr" gorm:"default:false"`
	Status         string    `json:"status" gorm:"type:varchar(20);default:enabled;index:idx_status"`
	EntryID        int       `json:"entryId" gorm:"default:0"`
}

// JobRecord 任务执行记录
type JobRecord struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CronjobID uint      `json:"cronjobId" gorm:"index:idx_cronjob_id"`
	StartTime int64     `json:"startTime"`
	Duration  int64     `json:"duration"`
	Status    string    `json:"status" gorm:"type:varchar(20);index:idx_record_status"`
	Message   string    `json:"message" gorm:"type:text"`
	LogFile   string    `json:"logFile" gorm:"type:varchar(500)"`
}
