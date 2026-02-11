package dto

// TerminalReq 终端请求
type TerminalReq struct {
	ID   string `json:"id"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// SSHConnectReq SSH 连接请求（通过 WebSocket 消息传输）
type SSHConnectReq struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password"` // 可选，如果使用密钥认证则为空
	Key      string `json:"key"`      // 可选，私钥内容
}