package controllers

import (
	"gpanel/global"
	"gpanel/utils"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket 升级器
var firewallUpGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		host := r.Host
		if origin == "" {
			return true
		}
		if strings.Contains(origin, host) {
			return true
		}
		return false
	},
}

// FirewallController 防火墙控制器（代理到Agent）
type FirewallController struct {
	agentClient *utils.AgentClient
}

func firewallError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"code": status, "message": message})
}

// NewFirewallController 创建防火墙控制器
func NewFirewallController() (*FirewallController, error) {
	client, err := utils.NewAgentClient()
	if err != nil {
		return nil, err
	}
	return &FirewallController{agentClient: client}, nil
}

// LoadBaseInfo 加载防火墙基础信息
func (c *FirewallController) LoadBaseInfo(ctx *gin.Context) {
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/base", nil)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// SearchRules 搜索防火墙规则
func (c *FirewallController) SearchRules(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		firewallError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/search", body)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// OperateFirewall 操作防火墙
func (c *FirewallController) OperateFirewall(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		firewallError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/operate", body)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// OperatePortRule 操作端口规则
func (c *FirewallController) OperatePortRule(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		firewallError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/port", body)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// UpdatePortRule 更新端口规则
func (c *FirewallController) UpdatePortRule(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		firewallError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/update/port", body)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// OperateIPRule 操作IP规则
func (c *FirewallController) OperateIPRule(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		firewallError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/ip", body)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// UpdateIPRule 更新IP规则
func (c *FirewallController) UpdateIPRule(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		firewallError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/update/ip", body)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// OperateForwardRule 操作端口转发规则
func (c *FirewallController) OperateForwardRule(ctx *gin.Context) {
	var body interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		firewallError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, statusCode, err := c.agentClient.RequestWithStatus(http.MethodPost, "/api/v1/firewall/forward", body)
	if err != nil {
		firewallError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Data(statusCode, "application/json", resp)
}

// getAgentWSBaseURL 获取 Agent WebSocket 基础 URL
func (c *FirewallController) getAgentWSBaseURL() string {
	agentAddr := "localhost:9998"
	if global.ConfigCacheInstance != nil {
		agentAddr = global.ConfigCacheInstance.GetAgentAddress()
	}
	return "ws://" + agentAddr
}

// InstallFirewall 安装防火墙 WebSocket 代理
// GET /api/v1/agent/firewall/install?type=ufw
func (c *FirewallController) InstallFirewall(ctx *gin.Context) {
	c.proxyWebSocket(ctx, c.getAgentWSBaseURL()+"/api/v1/firewall/install")
}

// UninstallFirewall 卸载防火墙 WebSocket 代理
// GET /api/v1/agent/firewall/uninstall?type=ufw&keepRules=false&keepPolicies=false
func (c *FirewallController) UninstallFirewall(ctx *gin.Context) {
	c.proxyWebSocket(ctx, c.getAgentWSBaseURL()+"/api/v1/firewall/uninstall")
}

// proxyWebSocket WebSocket 代理
func (c *FirewallController) proxyWebSocket(ctx *gin.Context, targetURL string) {
	// 升级到 WebSocket 连接
	clientConn, err := firewallUpGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("[ERROR] failed to upgrade websocket: %v", err)
		return
	}
	defer clientConn.Close()

	// 构建目标 URL
	query := ctx.Request.URL.Query()
	// 添加 API Key 认证（用于 Agent 端验证）
	if global.AgentAPIKey != "" {
		query.Set("api_key", global.AgentAPIKey)
	}
	targetURLWithQuery := targetURL + "?" + query.Encode()

	// 连接到 Agent 服务
	agentConn, _, err := websocket.DefaultDialer.Dial(targetURLWithQuery, nil)
	if err != nil {
		log.Printf("[ERROR] failed to connect to agent: %v", err)
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","message":"Failed to connect to agent service"}`))
		return
	}
	defer agentConn.Close()

	// 使用 done channel 来通知主 goroutine 退出
	done := make(chan struct{})
	var closeOnce sync.Once

	// 关闭 done channel 的安全函数
	closeDone := func() {
		closeOnce.Do(func() {
			close(done)
		})
	}

	// 客户端 -> Agent
	go func() {
		defer closeDone()
		for {
			select {
			case <-done:
				return
			default:
				messageType, message, err := clientConn.ReadMessage()
				if err != nil {
					return
				}
				if err := agentConn.WriteMessage(messageType, message); err != nil {
					return
				}
			}
		}
	}()

	// Agent -> 客户端
	go func() {
		defer closeDone()
		for {
			select {
			case <-done:
				return
			default:
				messageType, message, err := agentConn.ReadMessage()
				if err != nil {
					return
				}
				if err := clientConn.WriteMessage(messageType, message); err != nil {
					return
				}
			}
		}
	}()

	// 等待任一方向关闭
	<-done
}
