package controllers

import (
	"encoding/base64"
	"encoding/json"
	"gpanel/agent/dto"
	"gpanel/agent/repo"
	"gpanel/agent/utils/encrypt"
	"gpanel/agent/utils/ssh"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SettingController 设置控制器
type SettingController struct {
	settingRepo *repo.SettingRepo
}

func (sc *SettingController) success(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "success"
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": message, "data": data})
}

func (sc *SettingController) fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "message": message})
}

// NewSettingController 创建设置控制器
func NewSettingController() (*SettingController, error) {
	return &SettingController{settingRepo: repo.NewSettingRepo()}, nil
}

// SaveLocalConn 保存本地SSH连接配置
// POST /api/v1/settings/ssh
func (sc *SettingController) SaveLocalConn(c *gin.Context) {
	var req dto.LocalConnInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.fail(c, http.StatusBadRequest, err.Error())
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
		sc.fail(c, http.StatusBadRequest, err.Error())
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
	connAfterEncrypt, err := encrypt.StringEncrypt(string(localConn))
	if err != nil {
		sc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := sc.settingRepo.SaveOrUpdate("LocalSSHConn", connAfterEncrypt, "本地SSH连接配置"); err != nil {
		sc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	sc.success(c, nil, "Saved successfully")
}

// GetLocalConn 获取本地SSH连接配置
// GET /api/v1/settings/ssh/conn
func (sc *SettingController) GetLocalConn(c *gin.Context) {
	setting, err := sc.settingRepo.GetByKey("LocalSSHConn")
	if err != nil {
		sc.fail(c, http.StatusNotFound, "Local SSH connection not found")
		return
	}

	decrypted, err := encrypt.StringDecrypt(setting.Value)
	if err != nil {
		sc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	var connInfo dto.LocalConnInfo
	if err := json.Unmarshal([]byte(decrypted), &connInfo); err != nil {
		sc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	sc.success(c, connInfo, "success")
}

// TestLocalConn 测试本地SSH连接
// POST /api/v1/settings/ssh/check
func (sc *SettingController) TestLocalConn(c *gin.Context) {
	var req dto.LocalConnInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		sc.fail(c, http.StatusBadRequest, err.Error())
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
		sc.fail(c, http.StatusBadRequest, err.Error())
		return
	}
	defer client.Close()

	sc.success(c, gin.H{"success": true}, "Connection successful")
}
