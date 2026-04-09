package dto

type SSHInfo struct {
	IsActive        bool   `json:"isActive"`
	AutoStart       bool   `json:"autoStart"`
	Port            int    `json:"port"`
	ListenAddress   string `json:"listenAddress"`
	PasswordAuth    string `json:"passwordAuth"`
	PubkeyAuth      string `json:"pubkeyAuth"`
	PermitRootLogin string `json:"permitRootLogin"`
	UseDNS          string `json:"useDNS"`
}

type SSHSession struct {
	PID       int    `json:"pid"`
	Username  string `json:"username"`
	Terminal  string `json:"terminal"`
	Host      string `json:"host"`
	LoginTime string `json:"loginTime"`
}

type SSHLogItem struct {
	Address  string `json:"address"`
	Port     string `json:"port"`
	User     string `json:"user"`
	AuthMode string `json:"authMode"`
	Status   string `json:"status"`
	Date     string `json:"date"`
}

type SSHLogRes struct {
	Total int64        `json:"total"`
	Items []SSHLogItem `json:"items"`
}

type SSHConfigReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SSHOperateReq struct {
	Operation string `json:"operation"`
}

type SSHLogReq struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Status   string `json:"status"`
	Info     string `json:"info"`
}

type KillSessionReq struct {
	PID int `json:"pid"`
}

type SSHFileReq struct {
	Name string `json:"name"`
}

type SSHFileUpdateReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SSHKeyInfo struct {
	ID             uint   `json:"id"`
	CreatedAt      string `json:"createdAt"`
	Name           string `json:"name"`
	Mode           string `json:"mode"`
	EncryptionMode string `json:"encryptionMode"`
	PassPhrase     string `json:"passPhrase"`
	Description    string `json:"description"`
	PublicKey      string `json:"publicKey"`
	PrivateKey     string `json:"privateKey"`
}

type SSHKeyOperateReq struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Mode           string `json:"mode"`
	EncryptionMode string `json:"encryptionMode"`
	PassPhrase     string `json:"passPhrase"`
	Description    string `json:"description"`
	PublicKey      string `json:"publicKey"`
	PrivateKey     string `json:"privateKey"`
}

type SSHKeySearchReq struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type SSHKeyDeleteReq struct {
	IDs []uint `json:"ids"`
}

type SSHKeySearchRes struct {
	Total int64        `json:"total"`
	Items []SSHKeyInfo `json:"items"`
}
