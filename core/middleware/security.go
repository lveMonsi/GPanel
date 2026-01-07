package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Security 添加安全相关的 HTTP 响应头
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 启用浏览器内置的 XSS 过滤器
		c.Header("X-XSS-Protection", "1; mode=block")

		// 内容安全策略（CSP），防止混合内容
		// 根据请求路径动态调整 CSP 策略
		csp := buildCSP(c.Request.URL.Path)
		c.Header("Content-Security-Policy", csp)

		// 严格传输安全（HSTS），强制使用 HTTPS
		// 检查是否通过 HTTPS 访问（包括反向代理的情况）
		isHTTPS := c.Request.TLS != nil ||
			strings.ToLower(c.GetHeader("X-Forwarded-Proto")) == "https" ||
			strings.ToLower(c.GetHeader("X-Real-Proto")) == "https"

		if isHTTPS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// 推荐人策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 移除服务器信息泄露
		c.Header("Server", "GPanel")

		c.Next()
	}
}

// buildCSP 根据请求路径构建适当的 CSP 策略
func buildCSP(path string) string {
	// 基础 CSP 策略
	baseCSP := "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: https:; " +
		"font-src 'self' data:; " +
		"connect-src 'self'; " +
		"object-src 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self';"

	// API 路径可以使用更宽松的策略
	if strings.HasPrefix(path, "/api") {
		return "default-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none';"
	}

	return baseCSP
}