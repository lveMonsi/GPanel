package controllers

import (
	"gpanel/dto"
	"gpanel/global"
	"gpanel/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingController struct {
	settingService service.ISettingService
}

func settingError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "message": message})
}

func settingMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": message})
}

func NewSettingController() *SettingController {
	return &SettingController{
		settingService: service.NewSettingService(),
	}
}

func (sc *SettingController) GetAllSettings(c *gin.Context) {
	settings, err := sc.settingService.GetAllSettings()
	if err != nil {
		settingError(c, http.StatusInternalServerError, "Failed to get settings")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"settings": settings,
	})
}

func (sc *SettingController) GetSettingByKey(c *gin.Context) {
	key := c.Param("key")
	setting, err := sc.settingService.GetSettingByKey(key)
	if err != nil {
		settingError(c, http.StatusNotFound, "Setting not found")
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (sc *SettingController) UpdateSetting(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		settingError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := sc.settingService.UpdateSetting(req.Key, req.Value); err != nil {
		settingError(c, http.StatusInternalServerError, "Failed to update setting")
		return
	}

	// 更新缓存
	if global.ConfigCacheInstance != nil {
		global.ConfigCacheInstance.Set(req.Key, req.Value)
	}

	settingMessage(c, "Setting updated successfully")
}

func (sc *SettingController) CreateSetting(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
		About string `json:"about"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		settingError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := sc.settingService.CreateSetting(req.Key, req.Value, req.About); err != nil {
		settingError(c, http.StatusInternalServerError, "Failed to create setting")
		return
	}

	// 更新缓存
	if global.ConfigCacheInstance != nil {
		global.ConfigCacheInstance.Set(req.Key, req.Value)
	}

	settingMessage(c, "Setting created successfully")
}

func (sc *SettingController) DeleteSetting(c *gin.Context) {
	key := c.Param("key")

	if err := sc.settingService.DeleteSetting(key); err != nil {
		settingError(c, http.StatusInternalServerError, "Failed to delete setting")
		return
	}

	settingMessage(c, "Setting deleted successfully")
}

func (sc *SettingController) GetSystemSettings(c *gin.Context) {
	if global.ConfigCacheInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":     http.StatusOK,
			"settings": global.ConfigCacheInstance.GetAll(),
		})
	} else {
		settings, err := sc.settingService.GetAllSettings()
		if err != nil {
			settingError(c, http.StatusInternalServerError, "Failed to get system settings")
			return
		}

		// 将设置转换为键值对格式
		settingMap := make(map[string]string)
		for _, setting := range settings {
			settingMap[setting.Key] = setting.Value
		}

		c.JSON(http.StatusOK, gin.H{
			"code":     http.StatusOK,
			"settings": settingMap,
		})
	}
}

func (sc *SettingController) UpdateSystemSettings(c *gin.Context) {
	var req map[string]string

	if err := c.ShouldBindJSON(&req); err != nil {
		settingError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// 批量更新设置
	for key, value := range req {
		if err := sc.settingService.UpdateSetting(key, value); err != nil {
			settingError(c, http.StatusInternalServerError, "Failed to update setting: "+key)
			return
		}

		// 更新缓存
		if global.ConfigCacheInstance != nil {
			global.ConfigCacheInstance.Set(key, value)
		}
	}

	settingMessage(c, "System settings updated successfully")
}

// GetTerminalInfo 获取终端设置
// GET /api/v1/settings/terminal
func (sc *SettingController) GetTerminalInfo(c *gin.Context) {
	info, err := sc.settingService.GetTerminalInfo()
	if err != nil {
		settingError(c, http.StatusInternalServerError, "Failed to get terminal settings")
		return
	}

	c.JSON(http.StatusOK, info)
}

// UpdateTerminal 更新终端设置
// POST /api/v1/settings/terminal
func (sc *SettingController) UpdateTerminal(c *gin.Context) {
	var req dto.TerminalUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		settingError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := sc.settingService.UpdateTerminal(req); err != nil {
		settingError(c, http.StatusInternalServerError, "Failed to update terminal settings")
		return
	}

	settingMessage(c, "Terminal settings updated successfully")
}
