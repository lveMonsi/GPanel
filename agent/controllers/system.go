package controllers

import (
	"gpanel/agent/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemController struct{}

func NewSystemController() (*SystemController, error) {
	return &SystemController{}, nil
}

func (s *SystemController) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": data})
}

func (s *SystemController) fail(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
}

// GetSystemInfo 获取系统全量信息
func (s *SystemController) GetSystemInfo(c *gin.Context) {
	info, err := utils.GetSystemInfo()
	if err != nil {
		s.fail(c, err)
		return
	}
	s.success(c, info)
}

// GetCurrentInfo 获取当前实时信息
func (s *SystemController) GetCurrentInfo(c *gin.Context) {
	cpuInfo, memInfo, swapInfo, diskInfo, loadInfo, networkInfo, err := utils.GetCurrentInfo()
	if err != nil {
		s.fail(c, err)
		return
	}
	s.success(c, gin.H{
		"cpuInfo":     cpuInfo,
		"memoryInfo":  memInfo,
		"swapInfo":    swapInfo,
		"diskInfo":    diskInfo,
		"loadInfo":    loadInfo,
		"networkInfo": networkInfo,
	})
}

// GetOSInfo 获取操作系统信息
func (s *SystemController) GetOSInfo(c *gin.Context) {
	info := utils.GetOSInfo()
	s.success(c, gin.H{
		"os":   info.OS,
		"arch": info.Arch,
	})
}
