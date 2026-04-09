package controllers

import (
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SSHController struct {
	agentClient *utils.AgentClient
}

func sshError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

func NewSSHController() (*SSHController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &SSHController{agentClient: client}, nil
}

func (c *SSHController) proxy(ctx *gin.Context, method, path string) {
	var body interface{}
	if method == http.MethodPost {
		if err := ctx.ShouldBindJSON(&body); err != nil {
			sshError(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(method, path, body)
	if err != nil {
		sshError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

func (c *SSHController) GetSSHInfo(ctx *gin.Context) {
	c.proxy(ctx, http.MethodGet, "/api/v1/ssh/info")
}

func (c *SSHController) OperateSSH(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/operate")
}

func (c *SSHController) UpdateSSHConfig(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/config")
}

func (c *SSHController) LoadSSHFile(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/file")
}

func (c *SSHController) UpdateSSHFile(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/file/update")
}

func (c *SSHController) SearchSSHKeys(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/keys/search")
}

func (c *SSHController) CreateSSHKey(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/keys")
}

func (c *SSHController) UpdateSSHKey(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/keys/update")
}

func (c *SSHController) DeleteSSHKeys(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/keys/delete")
}

func (c *SSHController) SyncSSHKeys(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/keys/sync")
}

func (c *SSHController) GetSSHSessions(ctx *gin.Context) {
	c.proxy(ctx, http.MethodGet, "/api/v1/ssh/sessions")
}

func (c *SSHController) KillSSHSession(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/sessions/kill")
}

func (c *SSHController) GetSSHLogs(ctx *gin.Context) {
	c.proxy(ctx, http.MethodPost, "/api/v1/ssh/logs")
}
