package controllers

import (
	"gpanel/agent/dto"
	"gpanel/agent/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SSHController struct {
	sshService *service.SSHService
}

func NewSSHController() *SSHController {
	return &SSHController{sshService: service.NewSSHService()}
}

func (c *SSHController) success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": data})
}

func (c *SSHController) fail(ctx *gin.Context, status int, msg string) {
	ctx.JSON(status, gin.H{"code": status, "message": msg})
}

func (c *SSHController) GetSSHInfo(ctx *gin.Context) {
	info, err := c.sshService.GetSSHInfo()
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, info)
}

func (c *SSHController) OperateSSH(ctx *gin.Context) {
	var req dto.SSHOperateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.sshService.OperateSSH(req.Operation); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, nil)
}

func (c *SSHController) UpdateSSHConfig(ctx *gin.Context) {
	var req dto.SSHConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.sshService.UpdateSSHConfig(req.Key, req.Value); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, nil)
}

func (c *SSHController) GetSSHSessions(ctx *gin.Context) {
	sessions, err := c.sshService.GetSSHSessions()
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, sessions)
}

func (c *SSHController) KillSSHSession(ctx *gin.Context) {
	var req dto.KillSessionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.sshService.KillSSHSession(req.PID); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, nil)
}

func (c *SSHController) GetSSHLogs(ctx *gin.Context) {
	var req dto.SSHLogReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	result, err := c.sshService.GetSSHLogs(req.Page, req.PageSize, req.Status, req.Info)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.success(ctx, result)
}
