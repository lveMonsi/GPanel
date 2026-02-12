package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gpanel/agent/dto"
	"gpanel/agent/utils/ssh"
	"gpanel/agent/utils/terminal"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 16384,
	CheckOrigin: func(r *http.Request) bool {
		// 只允许同源或配置的域名
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
			"http://localhost:8080",
		}
		// 允许空 origin（某些浏览器或代理可能不发送）
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

// TerminalController 终端控制器
type TerminalController struct{}

// NewTerminalController 创建终端控制器
func NewTerminalController() (*TerminalController, error) {
	return &TerminalController{}, nil
}

// TerminalLocal 本地终端 WebSocket 端点
// GET /api/v1/terminal/local?cols=120&rows=30
func (tc *TerminalController) TerminalLocal(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("failed to upgrade websocket: %v", err)
		return
	}
	defer wsConn.Close()

	// 获取终端大小参数
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))

	// 创建本地命令
	slave, err := terminal.NewCommand("/bin/bash", "--login")
	if err != nil {
		tc.handleWsError(wsConn, err)
		return
	}
	defer slave.Close()

	// 创建 WebSocket 会话
	tty, err := terminal.NewLocalWsSession(cols, rows, wsConn, slave, true)
	if err != nil {
		tc.handleWsError(wsConn, err)
		return
	}

	// 启动会话
	quitChan := make(chan bool, 3)
	tty.Start(quitChan)
	go slave.Wait(quitChan)

	<-quitChan

	dt := time.Now().Add(time.Second)
	_ = wsConn.WriteControl(websocket.CloseMessage, nil, dt)
}

// TerminalSSH SSH 终端 WebSocket 端点
// GET /api/v1/terminal/ssh?cols=120&rows=30
// SSH 凭证通过 WebSocket 消息传输（第一条消息类型为 "connect"）
func (tc *TerminalController) TerminalSSH(c *gin.Context) {
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ERROR] failed to upgrade websocket: %v", err)
		return
	}
	defer wsConn.Close()

	// 获取终端大小参数并验证
	cols, err := strconv.Atoi(c.DefaultQuery("cols", "120"))
	if err != nil || cols < 10 || cols > 500 {
		cols = 120
	}
	rows, err := strconv.Atoi(c.DefaultQuery("rows", "30"))
	if err != nil || rows < 5 || rows > 200 {
		rows = 30
	}

	// 等待客户端发送连接消息（包含 SSH 凭证）
	var connectReq dto.SSHConnectReq
	_, wsData, err := wsConn.ReadMessage()
	if err != nil {
		tc.handleWsError(wsConn, err)
		return
	}

	msgObj := terminal.WsMsg{}
	if err := json.Unmarshal(wsData, &msgObj); err != nil {
		tc.handleWsError(wsConn, err)
		return
	}

	if msgObj.Type != terminal.WsMsgConnect {
		tc.handleWsError(wsConn, fmt.Errorf("expected connect message, got %s", msgObj.Type))
		return
	}

	// 解码 base64 数据
	decodeBytes, err := base64.StdEncoding.DecodeString(msgObj.Data)
	if err != nil {
		tc.handleWsError(wsConn, err)
		return
	}

	// 解析连接请求
	if err := json.Unmarshal(decodeBytes, &connectReq); err != nil {
		tc.handleWsError(wsConn, err)
		return
	}

	// 验证必要参数
	if connectReq.Host == "" {
		tc.handleWsError(wsConn, fmt.Errorf("host is required"))
		return
	}
	if connectReq.User == "" {
		tc.handleWsError(wsConn, fmt.Errorf("user is required"))
		return
	}
	if len(connectReq.User) > 64 {
		tc.handleWsError(wsConn, fmt.Errorf("username too long (max 64 characters)"))
		return
	}
	if len(connectReq.Password) > 1024 {
		tc.handleWsError(wsConn, fmt.Errorf("password too long (max 1024 characters)"))
		return
	}

	// 验证端口范围
	if connectReq.Port < 1 || connectReq.Port > 65535 {
		tc.handleWsError(wsConn, fmt.Errorf("port must be between 1 and 65535"))
		return
	}

	// 验证主机地址格式（简单的格式验证）
	if len(connectReq.Host) > 255 {
		tc.handleWsError(wsConn, fmt.Errorf("hostname too long (max 255 characters)"))
		return
	}

	// 验证认证模式
	if connectReq.Password == "" && connectReq.Key == "" {
		tc.handleWsError(wsConn, fmt.Errorf("either password or key is required"))
		return
	}

	// 创建 SSH 客户端
	connInfo := ssh.ConnInfo{
		User:        connectReq.User,
		Addr:        connectReq.Host,
		Port:        connectReq.Port,
		AuthMode:    "password",
		Password:    connectReq.Password,
		PrivateKey:  []byte(connectReq.Key),
		DialTimeOut: 5 * time.Second,
	}

	// 如果提供了密钥，则使用密钥认证
	if connectReq.Key != "" {
		connInfo.AuthMode = "key"
	}

	client, err := ssh.NewClient(connInfo)
	if err != nil {
		log.Printf("[ERROR] failed to create SSH client: %v", err)
		tc.handleWsError(wsConn, err)
		return
	}
	defer client.Close()

	// 创建 SSH 会话
	sws, err := terminal.NewLogicSshWsSession(cols, rows, client.Client, wsConn, "")
	if err != nil {
		log.Printf("[ERROR] failed to create SSH session: %v", err)
		tc.handleWsError(wsConn, err)
		return
	}
	defer sws.Close()

	// 启动会话
	quitChan := make(chan bool, 3)
	sws.Start(quitChan)
	go sws.Wait(quitChan)

	// 等待一小段时间，让 SSH 会话稳定
	time.Sleep(300 * time.Millisecond)

	// 发送连接成功消息（在会话启动后）
	successMsg := fmt.Sprintf("\r\n\x1b[32m[系统] SSH 连接成功！\x1b[m\r\n")
	wsData, jsonErr := json.Marshal(terminal.WsMsg{
		Type: terminal.WsMsgCmd,
		Data: base64.StdEncoding.EncodeToString([]byte(successMsg)),
	})
	if jsonErr != nil {
		log.Printf("[ERROR] failed to marshal success message: %v", jsonErr)
	} else {
		if writeErr := wsConn.WriteMessage(websocket.TextMessage, wsData); writeErr != nil {
			log.Printf("[WARN] failed to send success message: %v", writeErr)
		}
	}

	<-quitChan

	dt := time.Now().Add(time.Second)
	_ = wsConn.WriteControl(websocket.CloseMessage, nil, dt)
}

// handleWsError 处理 WebSocket 错误
func (tc *TerminalController) handleWsError(ws *websocket.Conn, err error) bool {
	if err != nil {
		log.Printf("[ERROR] terminal websocket error: %v", err)

		// 先发送错误消息到前端
		errorMsg := fmt.Sprintf("\r\n\x1b[31m[ERROR] %s\x1b[m\r\n", err.Error())
		wsData, jsonErr := json.Marshal(terminal.WsMsg{
			Type: terminal.WsMsgCmd,
			Data: base64.StdEncoding.EncodeToString([]byte(errorMsg)),
		})
		if jsonErr != nil {
			log.Printf("[ERROR] failed to marshal error message: %v", jsonErr)
			_ = ws.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"cmd\",\"data\":\"Connection error\"}"))
		} else {
			// 发送错误消息
			if writeErr := ws.WriteMessage(websocket.TextMessage, wsData); writeErr != nil {
				log.Printf("[WARN] failed to send error message: %v", writeErr)
			}
			// 等待消息发送
			time.Sleep(100 * time.Millisecond)
		}

		// 然后关闭连接
		dt := time.Now().Add(time.Second)
		closeMsg := []byte(err.Error())
		if ctlerr := ws.WriteControl(websocket.CloseMessage, closeMsg, dt); ctlerr != nil {
			log.Printf("[WARN] failed to send WebSocket close message: %v", ctlerr)
		}
		return true
	}
	return false
}