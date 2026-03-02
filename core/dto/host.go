package dto

// PageInfo 分页信息
type PageInfo struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// HostGroupOperate 主机分组操作
type HostGroupOperate struct {
	ID          uint   `json:"id"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// HostGroupSearch 主机分组搜索
type HostGroupSearch struct {
	PageInfo
	Info string `json:"info"` // 搜索关键词
}

// HostGroupInfo 主机分组信息
type HostGroupInfo struct {
	ID          uint   `json:"id"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HostCount   int    `json:"hostCount"` // 分组下的主机数量
}

// HostOperate 主机操作
type HostOperate struct {
	ID               uint   `json:"id"`
	GroupID          uint   `json:"groupID" binding:"required"`
	Name             string `json:"name" binding:"required"`
	Addr             string `json:"addr" binding:"required"`
	Port             int    `json:"port" binding:"required,min=1,max=65535"`
	User             string `json:"user" binding:"required"`
	AuthMode         string `json:"authMode" binding:"required,oneof=password key"`
	Password         string `json:"password"`
	PrivateKey       string `json:"privateKey"`
	PassPhrase       string `json:"passPhrase"`
	RememberPassword bool   `json:"rememberPassword"`
	Description      string `json:"description"`
}

// HostSearch 主机搜索
type HostSearch struct {
	PageInfo
	GroupID uint   `json:"groupID"` // 分组ID筛选
	Info    string `json:"info"`    // 搜索关键词（名称、地址、用户）
}

// HostInfo 主机信息
type HostInfo struct {
	ID               uint   `json:"id"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
	GroupID          uint   `json:"groupID"`
	GroupName        string `json:"groupName"` // 分组名称
	Name             string `json:"name"`
	Addr             string `json:"addr"`
	Port             int    `json:"port"`
	User             string `json:"user"`
	AuthMode         string `json:"authMode"`
	RememberPassword bool   `json:"rememberPassword"`
	Description      string `json:"description"`
	// 不返回密码和密钥
}

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

// HostTreeNode 主机树节点
type HostTreeNode struct {
	ID               uint            `json:"id"`
	Name             string          `json:"name"`
	Type             string          `json:"type"` // group 或 host
	GroupID          uint            `json:"groupID,omitempty"`
	Addr             string          `json:"addr,omitempty"`
	Port             int             `json:"port,omitempty"`
	User             string          `json:"user,omitempty"`
	AuthMode         string          `json:"authMode,omitempty"`
	RememberPassword bool            `json:"rememberPassword,omitempty"`
	Description      string          `json:"description,omitempty"`
	Children         []HostTreeNode  `json:"children"`
}

// HostMove 主机移动到其他分组
type HostMove struct {
	HostIDs []uint `json:"hostIDs" binding:"required"`
	GroupID uint   `json:"groupID" binding:"required"`
}