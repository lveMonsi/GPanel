package controllers

import (
	"gpanel/global"
	"gpanel/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UpdateController struct {
	service *service.UpdateService
}

func NewUpdateController() *UpdateController {
	return &UpdateController{service: service.NewUpdateService()}
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

func GetVersion(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "build": global.GetBuildInfo()})
}
