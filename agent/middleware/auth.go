package middleware

import (
	"gpanel/agent/global"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Auth API Key 认证中间件
// Core 通过 X-API-Key 头部或 api_key 查询参数（WebSocket 场景）传递 API Key
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := global.GetAPIKey()

		// 如果未配置 API Key，跳过认证（向后兼容）
		if apiKey == "" {
			c.Next()
			return
		}

		// 优先从 X-API-Key 头部获取
		key := c.GetHeader("X-API-Key")

		// 回退到查询参数（用于 WebSocket 升级请求）
		if key == "" {
			key = c.Query("api_key")
		}

		if key == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
			})
			c.Abort()
			return
		}

		if key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
