package controllers

import (
	"gpanel/utils"
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
	systemInfo, err := utils.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get system info",
		})
		return
	}
	c.JSON(http.StatusOK, systemInfo)
}

func GetCurrentInfo(c *gin.Context) {
	cpuInfo, memInfo, diskInfo, loadInfo, networkInfo, err := utils.GetCurrentInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get current info",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cpuInfo":     cpuInfo,
		"memoryInfo":  memInfo,
		"diskInfo":    diskInfo,
		"loadInfo":    loadInfo,
		"networkInfo": networkInfo,
	})
}