package repo

import (
	"errors"
	"gpanel/global"
	"gpanel/models"
	"gorm.io/gorm"
)

type SettingRepo struct{}

type ISettingRepo interface {
	List() ([]models.Setting, error)
	GetByKey(key string) (*models.Setting, error)
	GetValueByKey(key string) (string, error)
	Create(key, value, about string) error
	Update(key, value string) error
	UpdateOrCreate(key, value, about string) error
	Delete(key string) error
}

func NewSettingRepo() ISettingRepo {
	return &SettingRepo{}
}

func (r *SettingRepo) List() ([]models.Setting, error) {
	var settings []models.Setting
	err := global.DB.Find(&settings).Error
	return settings, err
}

func (r *SettingRepo) GetByKey(key string) (*models.Setting, error) {
	var setting models.Setting
	err := global.DB.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *SettingRepo) GetValueByKey(key string) (string, error) {
	var setting models.Setting
	if err := global.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *SettingRepo) Create(key, value, about string) error {
	setting := &models.Setting{
		Key:   key,
		Value: value,
		About: about,
	}
	return global.DB.Create(setting).Error
}

func (r *SettingRepo) Update(key, value string) error {
	return global.DB.Model(&models.Setting{}).Where("key = ?", key).Update("value", value).Error
}

func (r *SettingRepo) UpdateOrCreate(key, value, about string) error {
	var setting models.Setting
	result := global.DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return global.DB.Create(&models.Setting{Key: key, Value: value, About: about}).Error
		}
		return result.Error
	}
	return global.DB.Model(&setting).Updates(map[string]interface{}{"value": value, "about": about}).Error
}

func (r *SettingRepo) Delete(key string) error {
	return global.DB.Where("key = ?", key).Delete(&models.Setting{}).Error
}

// NewSettingRepoWithoutDB 创建不依赖 DB 的实例（用于避免循环导入）
func NewSettingRepoWithoutDB() *SettingRepo {
	return &SettingRepo{}
}