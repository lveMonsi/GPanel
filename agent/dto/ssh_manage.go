package dto

type SSHInfo struct {
	IsActive        bool   `json:"isActive"`
	Port            int    `json:"port"`
	ListenAddress   string `json:"listenAddress"`
	PasswordAuth    string `json:"passwordAuth"`
	PubkeyAuth      string `json:"pubkeyAuth"`
	PermitRootLogin string `json:"permitRootLogin"`
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
