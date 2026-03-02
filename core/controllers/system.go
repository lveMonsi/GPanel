package controllers

import (
	"gpanel/global"
	"gpanel/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "gpanel",
	})
}

// GetSystemInfo 获取系统信息
func GetSystemInfo(c *gin.Context) {
	info, err := utils.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetCurrentInfo 获取当前实时信息
func GetCurrentInfo(c *gin.Context) {
	cpuInfo, memInfo, diskInfo, loadInfo, networkInfo, err := utils.GetCurrentInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

// GetOSInfo 获取操作系统信息（用于前端检测）
func GetOSInfo(c *gin.Context) {
	info := utils.GetOSInfo()
	c.JSON(http.StatusOK, gin.H{
		"os":   info.OS,
		"arch": info.Arch,
	})
}

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	if global.ConfigCacheInstance == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config not initialized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"server": gin.H{
			"port":           global.ConfigCacheInstance.GetServerPort(),
			"mode":           global.ConfigCacheInstance.GetServerMode(),
			"sessionTimeout": global.ConfigCacheInstance.GetSessionTimeout(),
		},
		"security": gin.H{
			"entrance": global.ConfigCacheInstance.GetSecurityEntrance(),
		},
		"panel": gin.H{
			"user": global.ConfigCacheInstance.GetPanelUser(),
		},
		"system": gin.H{
			"language": global.ConfigCacheInstance.GetLanguage(),
			"timezone": global.ConfigCacheInstance.GetTimezone(),
		},
		"version": global.ConfigCacheInstance.GetVersion(),
	})
}

// UpdateConfig 更新配置
func UpdateConfig(c *gin.Context) {
	var req struct {
		ServerPort       string `json:"serverPort"`
		ServerMode       string `json:"serverMode"`
		SessionTimeout   int    `json:"sessionTimeout"`
		SecurityEntrance string `json:"securityEntrance"`
		Language         string `json:"language"`
		Timezone         string `json:"timezone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if global.ConfigCacheInstance == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config not initialized"})
		return
	}

	// 更新配置
	if req.ServerPort != "" {
		global.ConfigCacheInstance.UpdateSetting("ServerPort", req.ServerPort)
	}
	if req.ServerMode != "" {
		global.ConfigCacheInstance.UpdateSetting("ServerMode", req.ServerMode)
	}
	if req.SessionTimeout > 0 {
		global.ConfigCacheInstance.UpdateSetting("SessionTimeout", strconv.Itoa(req.SessionTimeout))
	}
	if req.SecurityEntrance != "" {
		global.ConfigCacheInstance.UpdateSetting("SecurityEntrance", req.SecurityEntrance)
	}
	if req.Language != "" {
		global.ConfigCacheInstance.UpdateSetting("Language", req.Language)
	}
	if req.Timezone != "" {
		global.ConfigCacheInstance.UpdateSetting("Timezone", req.Timezone)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated"})
}

// CheckConfigInitialized 检查配置是否已初始化
func CheckConfigInitialized(c *gin.Context) {
	if global.ConfigCacheInstance == nil {
		c.JSON(http.StatusOK, gin.H{"initialized": false})
		return
	}

	initialized := global.ConfigCacheInstance.IsInitialized()
	c.JSON(http.StatusOK, gin.H{"initialized": initialized})
}

// ReloadConfig 重新加载配置
func ReloadConfig(c *gin.Context) {
	if global.ConfigReloaderInstance == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config reloader not initialized"})
		return
	}

	// 触发配置重载
	if err := global.ConfigReloaderInstance.ReloadNow(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration reloaded"})
}