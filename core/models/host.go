package models

// HostGroup 主机分组模型
type HostGroup struct {
	BaseModel
	Name        string `json:"name" gorm:"type:varchar(128);not null;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`
}

// Host 主机模型
type Host struct {
	BaseModel
	GroupID          uint       `json:"groupID" gorm:"not null;index"`
	HostGroup        *HostGroup `json:"hostGroup,omitempty" gorm:"foreignKey:GroupID"`
	Name             string     `json:"name" gorm:"type:varchar(128);not null"`
	Addr             string     `json:"addr" gorm:"type:varchar(256);not null"`
	Port             int        `json:"port" gorm:"not null"`
	User             string     `json:"user" gorm:"type:varchar(64);not null"`
	AuthMode         string     `json:"authMode" gorm:"type:varchar(16);not null"` // password 或 key
	Password         string     `json:"password" gorm:"type:text"`                 // 加密存储
	PrivateKey       string     `json:"privateKey" gorm:"type:text"`               // 加密存储
	PassPhrase       string     `json:"passPhrase" gorm:"type:text"`               // 加密存储
	RememberPassword bool       `json:"rememberPassword" gorm:"default:false"`
	Description      string     `json:"description" gorm:"type:text"`
}