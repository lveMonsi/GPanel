package controllers

import (
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LogController 日志管理代理控制器（代理到 Agent）
type LogController struct {
	agentClient *utils.AgentClient
}

func NewLogController() (*LogController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &LogController{agentClient: client}, nil
}

func (c *LogController) proxyPost(ctx *gin.Context, path string) {
	var body any
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": err.Error()})
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, path, body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

func (c *LogController) proxyGet(ctx *gin.Context, path string) {
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodGet, path, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// SearchOperationLogs 搜索操作日志
func (c *LogController) SearchOperationLogs(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/logs/operation/search")
}

// CleanOperationLogs 清理操作日志
func (c *LogController) CleanOperationLogs(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/logs/operation/clean")
}

// GetOperationLogStats 获取操作日志统计
func (c *LogController) GetOperationLogStats(ctx *gin.Context) {
	c.proxyGet(ctx, "/api/v1/logs/operation/stats")
}

// SearchSystemLogs 搜索系统日志
func (c *LogController) SearchSystemLogs(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/logs/system/search")
}

// GetSystemLogInfo 获取系统日志文件信息
func (c *LogController) GetSystemLogInfo(ctx *gin.Context) {
	c.proxyGet(ctx, "/api/v1/logs/system/info")
}
