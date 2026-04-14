package controllers

import (
	"gpanel/agent/dto"
	"gpanel/agent/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogController struct {
	opLogSvc  *service.OperationLogService
	sysLogSvc *service.SystemLogService
}

func NewLogController() *LogController {
	return &LogController{
		opLogSvc:  service.NewOperationLogService(),
		sysLogSvc: service.NewSystemLogService(),
	}
}

func (c *LogController) ok(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": data})
}

func (c *LogController) fail(ctx *gin.Context, status int, msg string) {
	ctx.JSON(status, gin.H{"code": status, "message": msg})
}

// CreateOperationLog 创建操作日志
func (c *LogController) CreateOperationLog(ctx *gin.Context) {
	var req dto.OperationLogCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.opLogSvc.Create(req); err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, nil)
}

// SearchOperationLogs 搜索操作日志
func (c *LogController) SearchOperationLogs(ctx *gin.Context) {
	var req dto.OperationLogSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	logs, total, err := c.opLogSvc.Search(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, gin.H{"items": logs, "total": total})
}

// CleanOperationLogs 清理操作日志
func (c *LogController) CleanOperationLogs(ctx *gin.Context) {
	var req dto.OperationLogClean
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	deleted, err := c.opLogSvc.Clean(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, gin.H{"deleted": deleted})
}

// GetOperationLogStats 获取操作日志统计
func (c *LogController) GetOperationLogStats(ctx *gin.Context) {
	stats, err := c.opLogSvc.Stats()
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, stats)
}

// SearchSystemLogs 搜索系统日志
func (c *LogController) SearchSystemLogs(ctx *gin.Context) {
	var req dto.SystemLogSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.fail(ctx, http.StatusBadRequest, err.Error())
		return
	}
	lines, total, err := c.sysLogSvc.Search(req)
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, gin.H{"lines": lines, "total": total})
}

// GetSystemLogInfo 获取系统日志文件信息
func (c *LogController) GetSystemLogInfo(ctx *gin.Context) {
	info, err := c.sysLogSvc.GetInfo()
	if err != nil {
		c.fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.ok(ctx, info)
}
