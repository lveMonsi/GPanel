package dto

// SSHHost SSH 主机配置
type SSHHost struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Key      string `json:"key"`
}