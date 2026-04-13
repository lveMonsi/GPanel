package controllers

import (
	"net/http"
	"strconv"

	"gpanel/agent/dto"
	"gpanel/agent/service"

	"github.com/gin-gonic/gin"
)

type ProcessController struct {
	svc *service.ProcessService
}

func NewProcessController() *ProcessController {
	return &ProcessController{svc: service.NewProcessService()}
}

func (pc *ProcessController) ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": data})
}

func (pc *ProcessController) fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"code": status, "message": msg})
}

// ListProcesses 获取进程列表
func (pc *ProcessController) ListProcesses(c *gin.Context) {
	var req dto.ProcessSearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.fail(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := pc.svc.ListProcesses(req)
	if err != nil {
		pc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	pc.ok(c, data)
}

// GetProcessDetail 获取进程详情
func (pc *ProcessController) GetProcessDetail(c *gin.Context) {
	pidStr := c.Param("pid")
	pid, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		pc.fail(c, http.StatusBadRequest, "invalid pid")
		return
	}
	data, err := pc.svc.GetProcessDetail(int32(pid))
	if err != nil {
		pc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	pc.ok(c, data)
}

// StopProcess 终止进程
func (pc *ProcessController) StopProcess(c *gin.Context) {
	var req dto.ProcessStopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := pc.svc.StopProcess(req.PID); err != nil {
		pc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	pc.ok(c, nil)
}

// ListNetConnections 获取网络连接列表
func (pc *ProcessController) ListNetConnections(c *gin.Context) {
	var req dto.NetSearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.fail(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := pc.svc.ListNetConnections(req)
	if err != nil {
		pc.fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	pc.ok(c, data)
}
