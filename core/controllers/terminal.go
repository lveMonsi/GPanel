package controllers

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"gpanel/global"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 16384,
	CheckOrigin: func(r *http.Request) bool {
		// 允许同源请求
		origin := r.Header.Get("Origin")
		host := r.Host
		if origin == "" {
			return true
		}
		// 允许同源或配置的域名
		if strings.Contains(origin, host) {
			return true
		}
		return false
	},
}

// TerminalController 终端控制器（代理到 Agent 服务）
type TerminalController struct{}

// NewTerminalController 创建终端控制器
func NewTerminalController() (*TerminalController, error) {
	return &TerminalController{}, nil
}

// getAgentWSBaseURL 获取 Agent WebSocket 基础 URL
func (tc *TerminalController) getAgentWSBaseURL() string {
	agentAddr := "localhost:9998"
	if global.ConfigCacheInstance != nil {
		agentAddr = global.ConfigCacheInstance.GetAgentAddress()
	}
	return "ws://" + agentAddr
}

// TerminalLocal 本地终端 WebSocket 代理
// GET /api/v1/terminal/local?cols=120&rows=30
func (tc *TerminalController) TerminalLocal(c *gin.Context) {
	tc.proxyWebSocket(c, tc.getAgentWSBaseURL()+"/api/v1/terminal/local")
}

// TerminalSSH SSH 终端 WebSocket 代理
// GET /api/v1/terminal/ssh?cols=120&rows=30
func (tc *TerminalController) TerminalSSH(c *gin.Context) {
	tc.proxyWebSocket(c, tc.getAgentWSBaseURL()+"/api/v1/terminal/ssh")
}

// proxyWebSocket WebSocket 代理
func (tc *TerminalController) proxyWebSocket(c *gin.Context, targetURL string) {
	// 升级到 WebSocket 连接
	clientConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ERROR] failed to upgrade websocket: %v", err)
		return
	}
	defer clientConn.Close()

	// 构建目标 URL
	query := c.Request.URL.Query()
	// 添加 API Key 认证（用于 Agent 端验证）
	if global.AgentAPIKey != "" {
		query.Set("api_key", global.AgentAPIKey)
	}
	targetURLWithQuery := targetURL + "?" + query.Encode()

	// 连接到 Agent 服务
	agentConn, _, err := websocket.DefaultDialer.Dial(targetURLWithQuery, nil)
	if err != nil {
		log.Printf("[ERROR] failed to connect to agent: %v", err)
		// 向客户端发送错误消息
		clientConn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","data":"Failed to connect to agent service"}`))
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
