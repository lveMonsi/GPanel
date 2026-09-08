package controllers

import (
	"context"
	"errors"
	"gpanel/global"
	"gpanel/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UpdateController struct {
	service *service.UpdateService
}

func NewUpdateController() *UpdateController {
	return NewUpdateControllerWithService(service.NewUpdateService())
}

func NewUpdateControllerWithService(updateService *service.UpdateService) *UpdateController {
	return &UpdateController{service: updateService}
}

func (c *UpdateController) Start(ctx *gin.Context) {
	var req service.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": err.Error()})
		return
	}
	if err := c.service.Stage(req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "update already running" || err.Error() == "invalid version" {
			status = http.StatusConflict
		}
		ctx.JSON(status, gin.H{"code": status, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"code": http.StatusAccepted, "message": "update staging started", "status": c.service.Status()})
}

func (c *UpdateController) Apply(ctx *gin.Context) {
	if err := c.service.Apply(); err != nil {
		status := http.StatusConflict
		if err.Error() == "no staged update available" {
			status = http.StatusBadRequest
		}
		ctx.JSON(status, gin.H{"code": status, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"code": http.StatusAccepted, "message": "update apply started", "status": c.service.Status()})
}

func (c *UpdateController) Status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "status": c.service.Status()})
}

func (c *UpdateController) Latest(ctx *gin.Context) {
	channel := service.ReleaseChannel(ctx.DefaultQuery("channel", string(service.StableChannel)))
	if channel != service.StableChannel && channel != service.PrereleaseChannel {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "invalid release channel"})
		return
	}

	latest, err := c.service.Latest(ctx.Request.Context(), channel)
	if err != nil && !errors.Is(err, service.ErrNoRelease) {
		status := http.StatusBadGateway
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Request.Context().Err(), context.DeadlineExceeded) {
			status = http.StatusGatewayTimeout
		}
		ctx.JSON(status, gin.H{"code": status, "message": "获取远端版本失败"})
		return
	}
	current := global.GetBuildInfo()
	currentChannel := service.ClassifyReleaseChannel(current.Version)
	currentChannelValue := string(currentChannel)
	if currentChannelValue == "" {
		currentChannelValue = "unknown"
	}
	var latestResponse *service.LatestRelease
	if err == nil {
		latestResponse = &latest
	}
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "current": current, "currentChannel": currentChannelValue, "channel": channel, "latest": latestResponse})
}

func GetVersion(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "build": global.GetBuildInfo()})
}
