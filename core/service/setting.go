package service

import (
	"crypto/rand"
	"math/big"
	"time"

	"gpanel/dto"
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
	GetTerminalInfo() (*dto.TerminalInfo, error)
	UpdateTerminal(req dto.TerminalUpdate) error
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
		if err := settingRepo.Create(key, value, ""); err != nil {
			return err
		}
		return nil
	}
	if oldSetting.Value == value {
		return nil
	}
	if err := settingRepo.Update(key, value); err != nil {
		return err
	}
	return nil
}

func (s *SettingService) CreateSetting(key, value, about string) error {
	return settingRepo.Create(key, value, about)
}

func (s *SettingService) DeleteSetting(key string) error {
	return settingRepo.Delete(key)
}

func (s *SettingService) InitializeDefaultSettings() error {
	// 检查是否需要生成安全入口
	securityEntrance := "/"
	existingSetting, err := settingRepo.GetByKey("SecurityEntrance")
	if err == nil && existingSetting.Value != "/" {
		securityEntrance = existingSetting.Value
	} else {
		// 生成8位长度的小写字母和数字混合的安全入口
		securityEntrance = generateRandomEntrance()
	}

	defaultSettings := map[string]struct {
		Value string
		About string
	}{
		"ServerPort":               {"8080", "服务器端口"},
		"ServerMode":               {"debug", "服务器运行模式"},
		"SecurityEntrance":         {securityEntrance, "安全入口路径"},
		"Initialized":              {"true", "系统是否已初始化"},
		"Language":                 {"zh-CN", "系统语言"},
		"Timezone":                 {"Asia/Shanghai", "时区设置"},
		"PanelUser":                {"admin", "面板用户名"},
		"PanelPassword":            {"admin123", "面板密码"},
		"SessionTimeout":           {"86400", "会话超时时间（秒）"},
		"ServerAddress":            {"", "服务器地址"},
		"ListenAddress":            {"0.0.0.0", "监听地址"},
		"PasswordComplexityCheck":  {"false", "密码复杂度验证"},
		"AgentAddress":             {"localhost:9998", "Agent 服务地址"},
	}

	for key, setting := range defaultSettings {
		_, err := settingRepo.GetByKey(key)
		if err != nil {
			_ = settingRepo.Create(key, setting.Value, setting.About)
		}
	}

	return nil
}

// generateRandomEntrance 生成8位长度的小写字母和数字混合的安全入口
func generateRandomEntrance() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果随机数生成失败，返回一个默认的安全入口
			return "/secure" + time.Now().Format("20060102150405")
		}
		b[i] = charset[n.Int64()]
	}
	return "/" + string(b)
}

// GetTerminalInfo 获取终端设置
func (s *SettingService) GetTerminalInfo() (*dto.TerminalInfo, error) {
	settings, err := settingRepo.List()
	if err != nil {
		return nil, err
	}

	settingsMap := make(map[string]string)
	for _, s := range settings {
		settingsMap[s.Key] = s.Value
	}

	return &dto.TerminalInfo{
		LineHeight:        getSettingValue(settingsMap, "terminal_line_height", "1.2"),
		LetterSpacing:     getSettingValue(settingsMap, "terminal_letter_spacing", "1.2"),
		FontSize:          getSettingValue(settingsMap, "terminal_font_size", "14"),
		CursorBlink:       getSettingValue(settingsMap, "terminal_cursor_blink", "enable"),
		CursorStyle:       getSettingValue(settingsMap, "terminal_cursor_style", "underline"),
		Scrollback:        getSettingValue(settingsMap, "terminal_scrollback", "1000"),
		ScrollSensitivity: getSettingValue(settingsMap, "terminal_scroll_sensitivity", "10"),
	}, nil
}

// UpdateTerminal 更新终端设置
func (s *SettingService) UpdateTerminal(req dto.TerminalUpdate) error {
	settings := []models.Setting{
		{Key: "terminal_line_height", Value: req.LineHeight},
		{Key: "terminal_letter_spacing", Value: req.LetterSpacing},
		{Key: "terminal_font_size", Value: req.FontSize},
		{Key: "terminal_cursor_blink", Value: req.CursorBlink},
		{Key: "terminal_cursor_style", Value: req.CursorStyle},
		{Key: "terminal_scrollback", Value: req.Scrollback},
		{Key: "terminal_scroll_sensitivity", Value: req.ScrollSensitivity},
	}

	for _, setting := range settings {
		if err := s.UpdateSetting(setting.Key, setting.Value); err != nil {
			return err
		}
	}

	return nil
}

// getSettingValue 获取设置值，如果不存在则返回默认值
func getSettingValue(settings map[string]string, key, defaultValue string) string {
	if val, ok := settings[key]; ok && val != "" {
		return val
	}
	return defaultValue
}