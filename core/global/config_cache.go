package global

import (
	"log"
	"sync"
)

type ConfigCache struct {
	mu       sync.RWMutex
	settings map[string]string
}

var ConfigCacheInstance *ConfigCache

func InitConfigCache() error {
	ConfigCacheInstance = &ConfigCache{
		settings: make(map[string]string),
	}

	settingRepo := NewSettingRepo()
	settings, err := settingRepo.List()
	if err != nil {
		return err
	}

	for _, setting := range settings {
		ConfigCacheInstance.settings[setting.Key] = setting.Value
	}

	log.Printf("Config cache initialized with %d settings", len(ConfigCacheInstance.settings))
	return nil
}

func (cc *ConfigCache) Get(key string) (string, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	value, exists := cc.settings[key]
	return value, exists
}

func (cc *ConfigCache) Set(key, value string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.settings[key] = value
}

func (cc *ConfigCache) GetAll() map[string]string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range cc.settings {
		result[k] = v
	}
	return result
}

func (cc *ConfigCache) Reload() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	settingRepo := NewSettingRepo()
	settings, err := settingRepo.List()
	if err != nil {
		return err
	}

	cc.settings = make(map[string]string)
	for _, setting := range settings {
		cc.settings[setting.Key] = setting.Value
	}

	log.Printf("Config cache reloaded with %d settings", len(cc.settings))
	return nil
}

func (cc *ConfigCache) UpdateSetting(key, value string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	settingRepo := NewSettingRepo()
	if err := settingRepo.Update(key, value); err != nil {
		return err
	}

	cc.settings[key] = value
	return nil
}

func (cc *ConfigCache) GetServerPort() string {
	if port, exists := cc.Get("ServerPort"); exists {
		return port
	}
	return "8080"
}

func (cc *ConfigCache) GetServerMode() string {
	if mode, exists := cc.Get("ServerMode"); exists {
		return mode
	}
	return "debug"
}

func (cc *ConfigCache) GetSecurityEntrance() string {
	if entrance, exists := cc.Get("SecurityEntrance"); exists {
		return entrance
	}
	return "/"
}

func (cc *ConfigCache) GetLanguage() string {
	if lang, exists := cc.Get("Language"); exists {
		return lang
	}
	return "zh-CN"
}

func (cc *ConfigCache) GetTimezone() string {
	if tz, exists := cc.Get("Timezone"); exists {
		return tz
	}
	return "Asia/Shanghai"
}

func (cc *ConfigCache) IsInitialized() bool {
	initialized, exists := cc.Get("Initialized")
	return exists && initialized == "true"
}

// NewSettingRepo 创建 SettingRepo 实例（避免循环导入）
func NewSettingRepo() interface {
	List() ([]Setting, error)
	Update(key, value string) error
} {
	return &SettingRepo{}
}

// Setting 简化的 Setting 结构（避免循环导入）
type Setting struct {
	Key   string
	Value string
}

// SettingRepo 简化的 SettingRepo（避免循环导入）
type SettingRepo struct{}

func (r *SettingRepo) List() ([]Setting, error) {
	var settings []Setting
	err := DB.Table("settings").Select("key, value").Find(&settings).Error
	return settings, err
}

func (r *SettingRepo) Update(key, value string) error {
	return DB.Table("settings").Where("key = ?", key).Update("value", value).Error
}