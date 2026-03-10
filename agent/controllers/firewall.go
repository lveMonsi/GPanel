package controllers

import (
	"encoding/json"
	"gpanel/agent/dto"
	"gpanel/agent/global"
	"gpanel/agent/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket 升级器
var firewallUpGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := global.GetAllowedOrigins()
		origin := r.Header.Get("Origin")
		if len(allowedOrigins) == 0 {
			return true
		}
		if origin == "" {
			return true
		}
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// FirewallController 防火墙控制器
type FirewallController struct {
	firewallService *service.FirewallService
}

// NewFirewallController 创建防火墙控制器
func NewFirewallController() (*FirewallController, error) {
	return &FirewallController{
		firewallService: service.NewFirewallService(),
	}, nil
}

// LoadBaseInfo 加载防火墙基础信息
func (c *FirewallController) LoadBaseInfo(ctx *gin.Context) {
	info, err := c.firewallService.LoadBaseInfo()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": info})
}

// SearchRules 搜索防火墙规则
func (c *FirewallController) SearchRules(ctx *gin.Context) {
	var req dto.RuleSearch
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := c.firewallService.SearchRules(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// OperateFirewall 操作防火墙
func (c *FirewallController) OperateFirewall(ctx *gin.Context) {
	var req dto.FirewallOperation
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperateFirewall(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// OperatePortRule 操作端口规则
func (c *FirewallController) OperatePortRule(ctx *gin.Context) {
	var req dto.PortRuleOperate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperatePortRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// UpdatePortRule 更新端口规则
func (c *FirewallController) UpdatePortRule(ctx *gin.Context) {
	var req dto.PortRuleUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.UpdatePortRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// OperateIPRule 操作IP规则
func (c *FirewallController) OperateIPRule(ctx *gin.Context) {
	var req dto.IPRuleOperate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperateIPRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// UpdateIPRule 更新IP规则
func (c *FirewallController) UpdateIPRule(ctx *gin.Context) {
	var req dto.IPRuleUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.UpdateIPRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// OperateForwardRule 操作端口转发规则
func (c *FirewallController) OperateForwardRule(ctx *gin.Context) {
	var req dto.ForwardRuleOperate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.firewallService.OperateForwardRule(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// InstallFirewall 安装防火墙 (WebSocket)
// GET /api/v1/firewall/install?type=ufw
func (c *FirewallController) InstallFirewall(ctx *gin.Context) {
	// 升级为 WebSocket 连接
	wsConn, err := firewallUpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("failed to upgrade websocket: %v", err)
		return
	}
	defer wsConn.Close()

	// 获取防火墙类型参数
	firewallType := ctx.DefaultQuery("type", "ufw")

	// 创建进度回调函数
	progressChan := make(chan dto.InstallProgress, 100)

	// 启动安装服务
	go c.firewallService.InstallFirewall(firewallType, progressChan)

	// 发送进度消息到客户端
	for progress := range progressChan {
		data, err := json.Marshal(progress)
		if err != nil {
			log.Printf("failed to marshal progress: %v", err)
			continue
		}

		if err := wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("failed to write message: %v", err)
			break
		}

		// 如果安装完成或出错，关闭连接
		if progress.Type == "complete" || progress.Type == "error" {
			break
		}
	}
}