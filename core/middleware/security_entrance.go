package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"gpanel/global"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityEntrance 验证安全入口
func SecurityEntrance() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 跳过API路由和静态资源
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/assets") {
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

		// 如果安全入口是/，则允许所有路径，不校验sessionkey
		if securityEntrance == "/" {
			c.Next()
			return
		}

		// 检查请求路径是否匹配安全入口
		if path == securityEntrance {
			// 生成随机的 sessionkey
			sessionKey := generateSessionKey()

			// 设置 sessionkey cookie，30分钟过期
			// 使用 SameSite=Lax 允许跨站点导航时发送 cookie
			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie("sessionkey", sessionKey, 1800, "/", "", false, false)

			// 重定向到登录页面
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 对于其他所有路径，检查是否有有效的 sessionkey
		sessionKey, err := c.Cookie("sessionkey")
		if err != nil || sessionKey == "" {
			// 没有 sessionkey，返回404提示页面
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>暂时无法访问 - GPanel</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB',
                'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
            background: #F5F5F5;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }

        .not-found-card {
            background: #FFFFFF;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
            padding: 32px 40px;
            max-width: 500px;
            width: 100%;
            text-align: center;
        }

        .icon-wrapper {
            margin-bottom: 24px;
        }

        .lock-icon {
            width: 64px;
            height: 64px;
            color: #667eea;
        }

        .title {
            font-size: 22px;
            font-weight: 600;
            color: #333333;
            margin: 0 0 16px 0;
        }

        .description {
            font-size: 15px;
            color: #666666;
            line-height: 1.6;
            margin: 0 0 12px 0;
        }

        .instruction {
            font-size: 15px;
            color: #666666;
            line-height: 1.6;
            margin: 0 0 16px 0;
        }

        .code-block {
            display: inline-block;
            background: #EFEFEF;
            border-radius: 4px;
            padding: 8px 12px;
            margin-top: 16px;
        }

        .code-block code {
            font-family: 'Courier New', 'Consolas', 'Monaco', monospace;
            font-size: 14px;
            color: #333333;
            font-weight: 500;
        }

        @media (max-width: 480px) {
            .not-found-card {
                padding: 24px 20px;
            }

            .title {
                font-size: 20px;
            }

            .description,
            .instruction {
                font-size: 14px;
            }
        }
    </style>
</head>
<body>
    <div class="not-found-card">
        <div class="icon-wrapper">
            <svg class="lock-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 2C9.243 2 7 4.243 7 7V10H6C4.89543 10 4 10.8954 4 12V20C4 21.1046 4.89543 22 6 22H18C19.1046 22 20 21.1046 20 20V12C20 10.8954 19.1046 10 18 10H17V7C17 4.243 14.757 2 12 2ZM9 7C9 5.34315 10.3431 4 12 4C13.6569 4 15 5.34315 15 7V10H9V7ZM6 12H18V20H6V12Z" fill="#667eea"/>
            </svg>
        </div>
        <h1 class="title">暂时无法访问</h1>
        <p class="description">当前环境已经开启了安全入口登录</p>
        <p class="instruction">可在 SSH 终端输入以下命令来查看面板入口：</p>
        <div class="code-block">
            <code>gpctl user-info</code>
        </div>
    </div>
</body>
</html>`)
			c.Abort()
			return
		}

		// 有 sessionkey，放行
		c.Next()
	}
}

// generateSessionKey 生成随机的 sessionkey
func generateSessionKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}