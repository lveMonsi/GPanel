package controllers

import (
	"fmt"
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProcessController 进程管理代理控制器
type ProcessController struct {
	agentClient *utils.AgentClient
}

func processError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

func NewProcessController() (*ProcessController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &ProcessController{agentClient: client}, nil
}

func (c *ProcessController) proxyPost(ctx *gin.Context, path string) {
	var body any
	if err := ctx.ShouldBindJSON(&body); err != nil {
		processError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, path, body)
	if err != nil {
		processError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

func (c *ProcessController) proxyGet(ctx *gin.Context, path string) {
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodGet, path, nil)
	if err != nil {
		processError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// ListProcesses 获取进程列表
func (c *ProcessController) ListProcesses(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/process/list")
}

// GetProcessDetail 获取进程详情
func (c *ProcessController) GetProcessDetail(ctx *gin.Context) {
	pid := ctx.Param("pid")
	c.proxyGet(ctx, fmt.Sprintf("/api/v1/process/%s", pid))
}

// StopProcess 终止进程
func (c *ProcessController) StopProcess(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/process/stop")
}

// ListNetConnections 获取网络连接列表
func (c *ProcessController) ListNetConnections(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/process/net")
}
