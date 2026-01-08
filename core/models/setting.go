package models

type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"type:varchar(256);not null;uniqueIndex"`
	Value string `json:"value" gorm:"type:text"`
	About string `json:"about" gorm:"type:text"`
}