package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "GPanel API is running",
	})
}

func GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"hostname": "GPanel Server",
		"os":       "Linux",
		"version":  "1.0.0",
	})
}