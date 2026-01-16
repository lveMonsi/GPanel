package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gpanel/global"
)

var jwtSecret = []byte("gpanel-secret-key-change-in-production")

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 从配置缓存中获取用户名和密码
	panelUser := "admin"
	panelPassword := "admin123"
	if global.ConfigCacheInstance != nil {
		panelUser = global.ConfigCacheInstance.GetPanelUser()
		panelPassword = global.ConfigCacheInstance.GetPanelPassword()
	}

	// 验证用户名和密码
	if req.Username == panelUser && hashPassword(req.Password) == hashPassword(panelPassword) {
		// 获取会话超时时间（秒）
		sessionTimeout := 86400 // 默认 24 小时
		configVersion := ""
		if global.ConfigCacheInstance != nil {
			sessionTimeout = global.ConfigCacheInstance.GetSessionTimeout()
			configVersion = global.ConfigCacheInstance.GetVersion()
		}

		// 生成 JWT token，设置过期时间和配置版本
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username":       req.Username,
			"role":           "admin",
			"exp":            time.Now().Add(time.Duration(sessionTimeout) * time.Second).Unix(),
			"iat":            time.Now().Unix(),
			"config_version": configVersion,
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Invalid username or password",
	})
}