package controllers

import (
	"gpanel/global"
	"gpanel/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SystemController 系统信息控制器（代理到Agent）
type SystemController struct {
	agentClient *utils.AgentClient
}

func systemError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "message": message})
}

// NewSystemController 创建系统信息控制器
func NewSystemController() (*SystemController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &SystemController{agentClient: client}, nil
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "gpanel",
	})
}

// GetSystemInfo 获取系统信息（代理到Agent）
func (s *SystemController) GetSystemInfo(c *gin.Context) {
	resp, statusCode, err := s.agentClient.RequestWithStatus(http.MethodGet, "/api/v1/system/info", nil)
	if err != nil {
		systemError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(statusCode, "application/json", resp)
}

// GetCurrentInfo 获取当前实时信息（代理到Agent）
func (s *SystemController) GetCurrentInfo(c *gin.Context) {
	resp, statusCode, err := s.agentClient.RequestWithStatus(http.MethodGet, "/api/v1/system/current", nil)
	if err != nil {
		systemError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(statusCode, "application/json", resp)
}

// GetOSInfo 获取操作系统信息（代理到Agent）
func (s *SystemController) GetOSInfo(c *gin.Context) {
	resp, statusCode, err := s.agentClient.RequestWithStatus(http.MethodGet, "/api/v1/system/os", nil)
	if err != nil {
		systemError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(statusCode, "application/json", resp)
}

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	if global.ConfigCacheInstance == nil {
		systemError(c, http.StatusInternalServerError, "Config not initialized")
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
		systemError(c, http.StatusBadRequest, err.Error())
		return
	}

	if global.ConfigCacheInstance == nil {
		systemError(c, http.StatusInternalServerError, "Config not initialized")
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

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "Configuration updated"})
}

// CheckConfigInitialized 检查配置是否已初始化
func CheckConfigInitialized(c *gin.Context) {
	if global.ConfigCacheInstance == nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "initialized": false})
		return
	}

	initialized := global.ConfigCacheInstance.IsInitialized()
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "initialized": initialized})
}

// ReloadConfig 重新加载配置
func ReloadConfig(c *gin.Context) {
	if global.ConfigReloaderInstance == nil {
		systemError(c, http.StatusInternalServerError, "Config reloader not initialized")
		return
	}

	// 触发配置重载
	if err := global.ConfigReloaderInstance.ReloadNow(); err != nil {
		systemError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "Configuration reloaded"})
}
