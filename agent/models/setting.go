package models

import "time"

type BaseModel struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Setting 系统设置模型
type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"type:varchar(256);not null;uniqueIndex"`
	Value string `json:"value" gorm:"type:text"`
	About string `json:"about" gorm:"type:text"`
}

type SSHKey struct {
	BaseModel
	Name           string `json:"name" gorm:"type:varchar(128);not null;uniqueIndex"`
	Mode           string `json:"mode" gorm:"type:varchar(32);not null"`
	EncryptionMode string `json:"encryptionMode" gorm:"type:varchar(32);not null"`
	PassPhrase     string `json:"passPhrase" gorm:"type:text"`
	Description    string `json:"description" gorm:"type:text"`
	PublicKey      string `json:"publicKey" gorm:"type:text"`
	PrivateKey     string `json:"privateKey" gorm:"type:text"`
}
