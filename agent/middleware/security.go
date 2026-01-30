package middleware

import (
	"github.com/gin-gonic/gin"
)

// Security 中间件 - 添加安全相关的HTTP头
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 安全相关的HTTP头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}