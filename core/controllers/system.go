package controllers

import (
	"gpanel/global"
	"gpanel/service"
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "GPanel API is running",
	})
}

func GetSystemInfo(c *gin.Context) {
	systemInfo, err := utils.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get system info",
		})
		return
	}
	c.JSON(http.StatusOK, systemInfo)
}

func GetCurrentInfo(c *gin.Context) {
	cpuInfo, memInfo, diskInfo, loadInfo, networkInfo, err := utils.GetCurrentInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get current info",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cpuInfo":     cpuInfo,
		"memoryInfo":  memInfo,
		"diskInfo":    diskInfo,
		"loadInfo":    loadInfo,
		"networkInfo": networkInfo,
	})
}

func GetConfig(c *gin.Context) {
	if global.ConfigCacheInstance != nil {
		config := global.ConfigCacheInstance.GetAll()
		c.JSON(http.StatusOK, config)
	} else {
		settingService := service.NewSettingService()
		settings, err := settingService.GetAllSettings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get config",
			})
			return
		}

		configMap := make(map[string]string)
		for _, setting := range settings {
			configMap[setting.Key] = setting.Value
		}

		c.JSON(http.StatusOK, configMap)
	}
}

func UpdateConfig(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid config format",
		})
		return
	}

	settingService := service.NewSettingService()
	for key, value := range updates {
		if err := settingService.UpdateSetting(key, value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update config: " + key,
			})
			return
		}

		// 更新缓存
		if global.ConfigCacheInstance != nil {
			global.ConfigCacheInstance.Set(key, value)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config updated successfully",
	})
}

func CheckConfigInitialized(c *gin.Context) {
	if global.ConfigCacheInstance != nil {
		c.JSON(http.StatusOK, gin.H{
			"initialized": global.ConfigCacheInstance.IsInitialized(),
		})
	} else {
		settingService := service.NewSettingService()
		initialized, _ := settingService.GetSettingValueByKey("Initialized")
		c.JSON(http.StatusOK, gin.H{
			"initialized": initialized == "true",
		})
	}
}

func ReloadConfig(c *gin.Context) {
	if err := global.ReloadConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reload config",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config reloaded successfully",
	})
}