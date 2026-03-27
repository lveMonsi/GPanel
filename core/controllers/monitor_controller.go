package controllers

import (
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MonitorController 监控控制器（代理到 Agent）
type MonitorController struct {
	agentClient *utils.AgentClient
}

func monitorError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

// NewMonitorController 创建监控控制器
func NewMonitorController() (*MonitorController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &MonitorController{agentClient: client}, nil
}

// GetData 获取监控数据
func (c *MonitorController) GetData(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		monitorError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/monitor/data", body)
	if err != nil {
		monitorError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Data(statusCode, "application/json", resp)
}

// GetSetting 获取监控设置
func (c *MonitorController) GetSetting(ctx *gin.Context) {
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodGet, "/api/v1/monitor/setting", nil)
	if err != nil {
		monitorError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Data(statusCode, "application/json", resp)
}

// UpdateSetting 更新监控设置
func (c *MonitorController) UpdateSetting(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		monitorError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/monitor/setting", body)
	if err != nil {
		monitorError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Data(statusCode, "application/json", resp)
}

// ClearData 清空监控数据
func (c *MonitorController) ClearData(ctx *gin.Context) {
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodDelete, "/api/v1/monitor/data", nil)
	if err != nil {
		monitorError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Data(statusCode, "application/json", resp)
}
