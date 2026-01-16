package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gpanel/global"
)

var jwtSecret = []byte("gpanel-secret-key-change-in-production")

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// 检查 Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token expired, please login again",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
				})
			}
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("username", claims["username"])
			c.Set("role", claims["role"])

			// 检查过期时间
			if exp, ok := claims["exp"].(float64); ok {
				expTime := time.Unix(int64(exp), 0)
				if time.Now().After(expTime) {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "Token expired, please login again",
					})
					c.Abort()
					return
				}
			}

			// 检查配置版本，如果配置已变更则要求重新登录
			if tokenConfigVersion, ok := claims["config_version"].(string); ok {
				if global.ConfigCacheInstance != nil {
					currentConfigVersion := global.ConfigCacheInstance.GetVersion()
					if tokenConfigVersion != currentConfigVersion {
						c.JSON(http.StatusUnauthorized, gin.H{
							"error": "Configuration has been changed, please login again",
						})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}