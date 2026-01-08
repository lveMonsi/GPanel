package service

import (
	"gpanel/models"
	"gpanel/repo"
)

type SettingService struct{}

type ISettingService interface {
	GetAllSettings() ([]models.Setting, error)
	GetSettingByKey(key string) (*models.Setting, error)
	GetSettingValueByKey(key string) (string, error)
	UpdateSetting(key, value string) error
	CreateSetting(key, value, about string) error
	DeleteSetting(key string) error
	InitializeDefaultSettings() error
}

func NewSettingService() ISettingService {
	return &SettingService{}
}

var settingRepo repo.ISettingRepo = repo.NewSettingRepo()

func (s *SettingService) GetAllSettings() ([]models.Setting, error) {
	return settingRepo.List()
}

func (s *SettingService) GetSettingByKey(key string) (*models.Setting, error) {
	return settingRepo.GetByKey(key)
}

func (s *SettingService) GetSettingValueByKey(key string) (string, error) {
	return settingRepo.GetValueByKey(key)
}

func (s *SettingService) UpdateSetting(key, value string) error {
	oldSetting, err := settingRepo.GetByKey(key)
	if err != nil {
		// 如果设置不存在，则创建它
		_ = settingRepo.Create(key, value, "")
		return nil
	}
	if oldSetting.Value == value {
		return nil
	}
	return settingRepo.Update(key, value)
}

func (s *SettingService) CreateSetting(key, value, about string) error {
	return settingRepo.Create(key, value, about)
}

func (s *SettingService) DeleteSetting(key string) error {
	return settingRepo.Delete(key)
}

func (s *SettingService) InitializeDefaultSettings() error {
	defaultSettings := map[string]struct {
		Value string
		About string
	}{
		"ServerPort":      {"8080", "服务器端口"},
		"ServerMode":      {"debug", "服务器运行模式"},
		"SecurityEntrance": {"/", "安全入口路径"},
		"Initialized":     {"false", "系统是否已初始化"},
		"Language":        {"zh-CN", "系统语言"},
		"Timezone":        {"Asia/Shanghai", "时区设置"},
	}

	for key, setting := range defaultSettings {
		_, err := settingRepo.GetByKey(key)
		if err != nil {
			_ = settingRepo.Create(key, setting.Value, setting.About)
		}
	}

	return nil
}