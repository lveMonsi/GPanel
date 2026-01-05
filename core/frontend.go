package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path"
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

	r.GET("/", func(c *gin.Context) {
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.Request.URL.Path = path.Join("/assets", filepath)
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}