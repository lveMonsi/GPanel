package controllers

import (
	"gpanel/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MonitorController 监控控制器（代理到 Agent）
type MonitorController struct {
	agentClient *utils.AgentClient
}

type monitorQueryReq struct {
	Param     string `json:"param"`
	IO        string `json:"io,omitempty"`
	Network   string `json:"network,omitempty"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

func monitorError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

// NewMonitorController 创建监控控制器
func NewMonitorController() (*MonitorController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &MonitorController{agentClient: client}, nil
}

func (c *MonitorController) proxyToAgent(ctx *gin.Context, method, path string, body interface{}) {
	resp, statusCode, err := c.agentClient.RequestWithStatus(method, path, body)
	if err != nil {
		monitorError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Data(statusCode, "application/json", resp)
}

// QueryData 获取监控图表数据
func (c *MonitorController) QueryData(ctx *gin.Context) {
	monitorType := ctx.Param("type")
	switch monitorType {
	case "load", "cpu", "memory", "io", "network", "all":
	default:
		monitorError(ctx, http.StatusBadRequest, "unsupported monitor type")
		return
	}

	req := monitorQueryReq{
		Param:     monitorType,
		StartTime: ctx.Query("startTime"),
		EndTime:   ctx.Query("endTime"),
	}
	if monitorType == "io" {
		req.IO = ctx.DefaultQuery("device", "all")
	}
	if monitorType == "network" {
		req.Network = ctx.DefaultQuery("device", "all")
	}

	c.proxyToAgent(ctx, http.MethodPost, "/api/v1/monitor/query", req)
}

// GetData 获取监控数据
func (c *MonitorController) GetData(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		monitorError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	c.proxyToAgent(ctx, http.MethodPost, "/api/v1/monitor/query", body)
}

// GetIOOptions 获取磁盘设备选项
func (c *MonitorController) GetIOOptions(ctx *gin.Context) {
	c.proxyToAgent(ctx, http.MethodGet, "/api/v1/monitor/io-options", nil)
}

// GetNetworkOptions 获取网络设备选项
func (c *MonitorController) GetNetworkOptions(ctx *gin.Context) {
	c.proxyToAgent(ctx, http.MethodGet, "/api/v1/monitor/network-options", nil)
}

// GetSetting 获取监控设置
func (c *MonitorController) GetSetting(ctx *gin.Context) {
	c.proxyToAgent(ctx, http.MethodGet, "/api/v1/monitor/setting", nil)
}

// UpdateSetting 更新监控设置
func (c *MonitorController) UpdateSetting(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		monitorError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	c.proxyToAgent(ctx, http.MethodPost, "/api/v1/monitor/setting", body)
}

// ClearData 清空监控数据
func (c *MonitorController) ClearData(ctx *gin.Context) {
	c.proxyToAgent(ctx, http.MethodDelete, "/api/v1/monitor/data", nil)
}
