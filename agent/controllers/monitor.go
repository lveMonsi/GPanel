package controllers

import (
	"gpanel/agent/dto"
	"gpanel/agent/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MonitorController struct {
	service *service.MonitorService
}

func NewMonitorController() *MonitorController {
	return &MonitorController{
		service: service.NewMonitorService(),
	}
}

func (c *MonitorController) QueryData(ctx *gin.Context) {
	var req dto.MonitorQueryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := c.service.QueryData(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": data})
}

func (c *MonitorController) GetIOOptions(ctx *gin.Context) {
	options, err := c.service.GetIOOptions()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": options})
}

func (c *MonitorController) GetNetworkOptions(ctx *gin.Context) {
	options, err := c.service.GetNetworkOptions()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": options})
}

func (c *MonitorController) GetSetting(ctx *gin.Context) {
	setting, err := c.service.GetSetting()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": setting})
}

func (c *MonitorController) UpdateSetting(ctx *gin.Context) {
	var req dto.MonitorSettingReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.UpdateSetting(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "设置已更新"})
}
