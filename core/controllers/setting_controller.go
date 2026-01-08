package controllers

import (
	"gpanel/global"
	"gpanel/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingController struct {
	settingService service.ISettingService
}

func NewSettingController() *SettingController {
	return &SettingController{
		settingService: service.NewSettingService(),
	}
}

func (sc *SettingController) GetAllSettings(c *gin.Context) {
	settings, err := sc.settingService.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get settings",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

func (sc *SettingController) GetSettingByKey(c *gin.Context) {
	key := c.Param("key")
	setting, err := sc.settingService.GetSettingByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Setting not found",
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if err := sc.settingService.UpdateSetting(req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update setting",
		})
		return
	}

	// 更新缓存
	if global.ConfigCacheInstance != nil {
		global.ConfigCacheInstance.Set(req.Key, req.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting updated successfully",
	})
}

func (sc *SettingController) CreateSetting(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
		About string `json:"about"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if err := sc.settingService.CreateSetting(req.Key, req.Value, req.About); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create setting",
		})
		return
	}

	// 更新缓存
	if global.ConfigCacheInstance != nil {
		global.ConfigCacheInstance.Set(req.Key, req.Value)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting created successfully",
	})
}

func (sc *SettingController) DeleteSetting(c *gin.Context) {
	key := c.Param("key")

	if err := sc.settingService.DeleteSetting(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete setting",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting deleted successfully",
	})
}

func (sc *SettingController) GetSystemSettings(c *gin.Context) {
	if global.ConfigCacheInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"settings": global.ConfigCacheInstance.GetAll(),
		})
	} else {
		settings, err := sc.settingService.GetAllSettings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get system settings",
			})
			return
		}

		// 将设置转换为键值对格式
		settingMap := make(map[string]string)
		for _, setting := range settings {
			settingMap[setting.Key] = setting.Value
		}

		c.JSON(http.StatusOK, gin.H{
			"settings": settingMap,
		})
	}
}

func (sc *SettingController) UpdateSystemSettings(c *gin.Context) {
	var req map[string]string

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 批量更新设置
	for key, value := range req {
		if err := sc.settingService.UpdateSetting(key, value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update setting: " + key,
			})
			return
		}

		// 更新缓存
		if global.ConfigCacheInstance != nil {
			global.ConfigCacheInstance.Set(key, value)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "System settings updated successfully",
	})
}