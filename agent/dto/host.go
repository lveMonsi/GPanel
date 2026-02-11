package dto

// HostConnTest 主机连接测试
type HostConnTest struct {
	Addr       string `json:"addr" binding:"required"`
	Port       int    `json:"port" binding:"required,min=1,max=65535"`
	User       string `json:"user" binding:"required"`
	AuthMode   string `json:"authMode" binding:"required,oneof=password key"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	PassPhrase string `json:"passPhrase"`
}