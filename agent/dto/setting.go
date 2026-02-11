package dto

// LocalConnInfo 本地SSH连接信息
type LocalConnInfo struct {
	Addr       string `json:"addr"`
	Port       uint   `json:"port"`
	User       string `json:"user"`
	AuthMode   string `json:"authMode"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	PassPhrase string `json:"passPhrase"`
}