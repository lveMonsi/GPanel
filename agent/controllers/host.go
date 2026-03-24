package controllers

import (
	"gpanel/agent/dto"
	"gpanel/agent/utils/ssh"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HostController 主机控制器
type HostController struct{}

func (hc *HostController) success(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "success"
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": message, "data": data})
}

func (hc *HostController) fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "message": message})
}

// NewHostController 创建主机控制器
func NewHostController() (*HostController, error) {
	return &HostController{}, nil
}

// TestConnection 测试主机连接
// POST /api/v1/hosts/test
func (hc *HostController) TestConnection(c *gin.Context) {
	var req dto.HostConnTest
	if err := c.ShouldBindJSON(&req); err != nil {
		hc.fail(c, http.StatusBadRequest, err.Error())
		return
	}

	// 创建 SSH 连接信息
	connInfo := ssh.ConnInfo{
		User:        req.User,
		Addr:        req.Addr,
		Port:        req.Port,
		AuthMode:    req.AuthMode,
		DialTimeOut: 5 * time.Second,
	}

	if req.AuthMode == "password" {
		connInfo.Password = req.Password
	} else {
		connInfo.PrivateKey = []byte(req.PrivateKey)
		if req.PassPhrase != "" {
			connInfo.PassPhrase = []byte(req.PassPhrase)
		}
	}

	// 尝试连接
	client, err := ssh.NewClient(connInfo)
	if err != nil {
		log.Printf("[ERROR] SSH connection test failed: %v", err)
		hc.fail(c, http.StatusBadRequest, err.Error())
		return
	}
	defer client.Close()

	hc.success(c, gin.H{"success": true}, "Connection successful")
}
