package controllers

import (
	"gpanel/agent/dto"
	"gpanel/agent/utils/ssh"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// HostController 主机控制器
type HostController struct{}

// NewHostController 创建主机控制器
func NewHostController() (*HostController, error) {
	return &HostController{}, nil
}

// TestConnection 测试主机连接
// POST /api/v1/hosts/test
func (hc *HostController) TestConnection(c *gin.Context) {
	var req dto.HostConnTest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
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
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer client.Close()

	c.JSON(200, gin.H{"success": true, "message": "Connection successful"})
}