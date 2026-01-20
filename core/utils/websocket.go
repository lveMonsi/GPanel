package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"gpanel/dto"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// WebSocketClient WebSocket 客户端
type WebSocketClient struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *WebSocketHub
	UserID string
}

// WebSocketMessage WebSocket 消息
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
}

// WebSocketHub WebSocket 集线器
type WebSocketHub struct {
	clients    map[string]*WebSocketClient
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan *WebSocketMessage
	mu         sync.RWMutex
}

// NewWebSocketHub 创建 WebSocket 集线器
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan *WebSocketMessage),
	}
}

// Run 运行 WebSocket 集线器
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("WebSocket client registered: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered: %s", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- h.encodeMessage(message):
				default:
					// 发送失败，关闭连接
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register 注册客户端
func (h *WebSocketHub) Register(client *WebSocketClient) {
	h.register <- client
}

// Unregister 注销客户端
func (h *WebSocketHub) Unregister(client *WebSocketClient) {
	h.unregister <- client
}

// Broadcast 广播消息
func (h *WebSocketHub) Broadcast(message *WebSocketMessage) {
	h.broadcast <- message
}

// SendToClient 发送消息给指定客户端
func (h *WebSocketHub) SendToClient(clientID string, message *WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if client, ok := h.clients[clientID]; ok {
		select {
		case client.Send <- h.encodeMessage(message):
		default:
			// 发送失败
		}
	}
}

// encodeMessage 编码消息
func (h *WebSocketHub) encodeMessage(message *WebSocketMessage) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to encode WebSocket message: %v", err)
		return nil
	}
	return data
}

// WritePump 写入消息到客户端
func (c *WebSocketClient) WritePump() {
	defer c.Conn.Close()
	
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to write WebSocket message: %v", err)
				return
			}
		}
	}
}

// ReadPump 从客户端读取消息
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(512)
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
		
		// 处理客户端消息
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Failed to unmarshal WebSocket message: %v", err)
			continue
		}
		
		// 这里可以根据消息类型处理不同的逻辑
		// 例如：心跳检测、订阅特定频道等
	}
}

// 全局 WebSocket 集线器
var Hub = NewWebSocketHub()

// HandleWebSocket 处理 WebSocket 连接
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	
	// 从查询参数获取客户端 ID
	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = uuid.New().String()
	}
	
	// 从 JWT 获取用户 ID（如果有）
	userID := ""
	if userIDValue, exists := c.Get("user_id"); exists {
		if uid, ok := userIDValue.(string); ok {
			userID = uid
		}
	}
	
	client := &WebSocketClient{
		ID:     clientID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    Hub,
		UserID: userID,
	}
	
	Hub.Register(client)
	
	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}

// BroadcastProgress 广播进度更新
func BroadcastProgress(progressKey string, progress *dto.ProgressInfo) {
	message := &WebSocketMessage{
		Type: "progress",
		Data: map[string]interface{}{
			"key":     progressKey,
			"progress": progress,
		},
	}
	Hub.Broadcast(message)
}

// SendProgressToClient 发送进度更新给指定客户端
func SendProgressToClient(clientID string, progressKey string, progress *dto.ProgressInfo) {
	message := &WebSocketMessage{
		Type: "progress",
		Data: map[string]interface{}{
			"key":     progressKey,
			"progress": progress,
		},
	}
	Hub.SendToClient(clientID, message)
}