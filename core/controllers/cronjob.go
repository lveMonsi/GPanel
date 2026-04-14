package controllers

import (
	"fmt"
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CronjobController 计划任务代理控制器（代理到 Agent）
type CronjobController struct {
	agentClient *utils.AgentClient
}

func NewCronjobController() (*CronjobController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &CronjobController{agentClient: client}, nil
}

func (c *CronjobController) proxyPost(ctx *gin.Context, path string) {
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

func (c *CronjobController) proxyGet(ctx *gin.Context, path string) {
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodGet, path, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

func (c *CronjobController) Create(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs")
}

func (c *CronjobController) Update(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/update")
}

func (c *CronjobController) Delete(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/delete")
}

func (c *CronjobController) Search(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/search")
}

func (c *CronjobController) Toggle(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/toggle")
}

func (c *CronjobController) HandleOnce(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/handle")
}

func (c *CronjobController) StopRunning(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/stop")
}

func (c *CronjobController) SearchRecords(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/records/search")
}

func (c *CronjobController) GetRecordLog(ctx *gin.Context) {
	id := ctx.Param("id")
	c.proxyGet(ctx, fmt.Sprintf("/api/v1/cronjobs/records/%s/log", id))
}

func (c *CronjobController) CleanRecords(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/records/clean")
}

func (c *CronjobController) GetNextExecTimes(ctx *gin.Context) {
	c.proxyPost(ctx, "/api/v1/cronjobs/next-times")
}
