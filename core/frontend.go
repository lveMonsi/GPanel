package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var frontendFS embed.FS

func SetupFrontend(r *gin.Engine) {
	frontendDist, err := fs.Sub(frontendFS, "web/dist")
	if err != nil {
		log.Printf("Warning: Failed to load frontend files: %v", err)
		return
	}

	fileServer := http.FileServer(http.FS(frontendDist))

	// 处理静态资源
	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.Request.URL.Path = "/assets/" + filepath
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// 使用 NoRoute 处理所有其他路由（包括前端路由）
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是 API 请求，返回 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// 对于其他所有请求，返回 index.html 以支持前端路由
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}