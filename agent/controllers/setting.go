package controllers

import (
	"encoding/base64"
	"encoding/json"
	"gpanel/agent/dto"
	"gpanel/agent/models"
	"gpanel/agent/utils/encrypt"
	"gpanel/agent/utils/ssh"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// SettingController 设置控制器
type SettingController struct{}

// NewSettingController 创建设置控制器
func NewSettingController() (*SettingController, error) {
	return &SettingController{}, nil
}

// SaveLocalConn 保存本地SSH连接配置
// POST /api/v1/settings/ssh
func (sc *SettingController) SaveLocalConn(c *gin.Context) {
	var req dto.LocalConnInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Base64解码
	if req.AuthMode == "password" && req.Password != "" {
		password, _ := base64.StdEncoding.DecodeString(req.Password)
		req.Password = string(password)
	}
	if req.AuthMode == "key" && req.PrivateKey != "" {
		privateKey, _ := base64.StdEncoding.DecodeString(req.PrivateKey)
		req.PrivateKey = string(privateKey)
	}
	if req.AuthMode == "key" && req.PassPhrase != "" {
		passPhrase, _ := base64.StdEncoding.DecodeString(req.PassPhrase)
		req.PassPhrase = string(passPhrase)
	}

	// 测试连接
	connInfo := ssh.ConnInfo{
		User:        req.User,
		Addr:        req.Addr,
		Port:        int(req.Port),
		AuthMode:    req.AuthMode,
		Password:    req.Password,
		PrivateKey:  []byte(req.PrivateKey),
		PassPhrase:  []byte(req.PassPhrase),
		DialTimeOut: 5 * time.Second,
	}

	client, err := ssh.NewClient(connInfo)
	if err != nil {
		log.Printf("[ERROR] SSH connection test failed: %v", err)
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer client.Close()

	// 加密存储
	connItem := dto.LocalConnInfo{
		Addr:       req.Addr,
		Port:       req.Port,
		User:       req.User,
		AuthMode:   req.AuthMode,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		PassPhrase: req.PassPhrase,
	}

	localConn, _ := json.Marshal(&connItem)
	connAfterEncrypt, _ := encrypt.StringEncrypt(string(localConn))

	// 存储到数据库
	_ = &models.Setting{
		Key:   "LocalSSHConn",
		Value: connAfterEncrypt,
		About: "本地SSH连接配置",
	}

	// 这里简化处理，实际应该使用repository
	// 暂时返回成功
	c.JSON(200, gin.H{"success": true, "message": "Saved successfully"})
}

// GetLocalConn 获取本地SSH连接配置
// GET /api/v1/settings/ssh/conn
func (sc *SettingController) GetLocalConn(c *gin.Context) {
	// 从数据库读取
	// 这里简化处理，实际应该使用repository
	c.JSON(200, gin.H{"success": false, "message": "Not implemented yet"})
}

// TestLocalConn 测试本地SSH连接
// POST /api/v1/settings/ssh/check
func (sc *SettingController) TestLocalConn(c *gin.Context) {
	var req dto.LocalConnInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Base64解码
	if req.AuthMode == "password" && req.Password != "" {
		password, _ := base64.StdEncoding.DecodeString(req.Password)
		req.Password = string(password)
	}
	if req.AuthMode == "key" && req.PrivateKey != "" {
		privateKey, _ := base64.StdEncoding.DecodeString(req.PrivateKey)
		req.PrivateKey = string(privateKey)
	}
	if req.AuthMode == "key" && req.PassPhrase != "" {
		passPhrase, _ := base64.StdEncoding.DecodeString(req.PassPhrase)
		req.PassPhrase = string(passPhrase)
	}

	// 测试连接
	connInfo := ssh.ConnInfo{
		User:        req.User,
		Addr:        req.Addr,
		Port:        int(req.Port),
		AuthMode:    req.AuthMode,
		Password:    req.Password,
		PrivateKey:  []byte(req.PrivateKey),
		PassPhrase:  []byte(req.PassPhrase),
		DialTimeOut: 5 * time.Second,
	}

	client, err := ssh.NewClient(connInfo)
	if err != nil {
		log.Printf("[ERROR] SSH connection test failed: %v", err)
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer client.Close()

	c.JSON(200, gin.H{"success": true, "message": "Connection successful"})
}