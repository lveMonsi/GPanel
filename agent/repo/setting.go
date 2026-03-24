package repo

import (
	"errors"

	"gpanel/agent/global"
	"gpanel/agent/models"

	"gorm.io/gorm"
)

type SettingRepo struct{}

func NewSettingRepo() *SettingRepo {
	return &SettingRepo{}
}

func (r *SettingRepo) SaveOrUpdate(key, value, about string) error {
	var setting models.Setting
	err := global.DB.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return global.DB.Create(&models.Setting{Key: key, Value: value, About: about}).Error
		}
		return err
	}
	return global.DB.Model(&setting).Updates(map[string]interface{}{"value": value, "about": about}).Error
}

func (r *SettingRepo) GetByKey(key string) (*models.Setting, error) {
	var setting models.Setting
	if err := global.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}
