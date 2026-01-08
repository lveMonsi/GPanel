package middleware

import (
	"gpanel/global"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityEntrance 验证安全入口
func SecurityEntrance() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 跳过API路由和静态资源
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/static") {
			c.Next()
			return
		}

		// 获取配置的安全入口
		var securityEntrance string
		if global.ConfigCacheInstance != nil {
			securityEntrance = global.ConfigCacheInstance.GetSecurityEntrance()
		} else {
			securityEntrance = "/"
		}

		// 确保安全入口以/开头
		if !strings.HasPrefix(securityEntrance, "/") {
			securityEntrance = "/" + securityEntrance
		}

		// 检查请求路径是否匹配安全入口
		if path == securityEntrance || strings.HasPrefix(path, securityEntrance+"/") {
			c.Next()
			return
		}

		// 如果安全入口是/，则允许所有路径
		if securityEntrance == "/" {
			c.Next()
			return
		}

		// 不匹配则返回404
		c.AbortWithStatus(404)
	}
}