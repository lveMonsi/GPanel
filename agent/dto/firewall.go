package dto

// FirewallBaseInfo 防火墙基础信息
type FirewallBaseInfo struct {
	Name       string `json:"name"`       // 防火墙名称 (ufw, iptables)
	IsExist    bool   `json:"isExist"`    // 是否存在
	IsActive   bool   `json:"isActive"`   // 是否激活
	IsInit     bool   `json:"isInit"`     // 是否初始化
	IsBind     bool   `json:"isBind"`     // 是否绑定
	Version    string `json:"version"`    // 版本
	PingStatus string `json:"pingStatus"` // Ping 状态 (enabled/disabled)
}

// PageInfo 分页信息
type PageInfo struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// RuleSearch 规则搜索
type RuleSearch struct {
	PageInfo
	Info     string `json:"info"`     // 搜索信息（IP或端口）
	Status   string `json:"status"`   // 状态筛选
	Strategy string `json:"strategy"` // 策略筛选
	Type     string `json:"type"`     // 类型
}

// FireInfo 防火墙规则信息
type FireInfo struct {
	ID          uint   `json:"id"`
	Protocol    string `json:"protocol"`    // 协议
	Port        string `json:"port"`        // 端口
	Address     string `json:"address"`     // IP地址
	Strategy    string `json:"strategy"`    // 策略
	UsedStatus  string `json:"usedStatus"`  // 使用状态
	Description string `json:"description"` // 描述
	SourcePort  string `json:"sourcePort"`  // 源端口
	TargetPort  string `json:"targetPort"`  // 目标端口
	TargetIP    string `json:"targetIP"`    // 目标IP
	Chain       string `json:"chain"`       // 链名称
	Family      string `json:"family"`      // IP家族 (ipv4/ipv6)
	Num         string `json:"num"`         // 规则编号（用于更新/删除）
	Interface   string `json:"interface"`   // 网络接口
}

// FirewallOperation 防火墙操作
type FirewallOperation struct {
	Operation         string `json:"operation"`         // 操作类型: start, stop, restart, enableBanPing, disableBanPing
	WithDockerRestart bool   `json:"withDockerRestart"` // 是否重启Docker
}

// PortRuleOperate 端口规则操作
type PortRuleOperate struct {
	Operation   string `json:"operation" validate:"required,oneof=add remove"`
	Port        string `json:"port" validate:"required"`
	Protocol    string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
	Strategy    string `json:"strategy" validate:"required,oneof=accept drop"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// PortRuleUpdate 端口规则更新
type PortRuleUpdate struct {
	OldRule PortRuleOperate `json:"oldRule"`
	NewRule PortRuleOperate `json:"newRule"`
}

// IPRuleOperate IP规则操作
type IPRuleOperate struct {
	Operation   string `json:"operation" validate:"required,oneof=add remove"`
	Address     string `json:"address" validate:"required"`
	Strategy    string `json:"strategy" validate:"required,oneof=accept drop"`
	Protocol    string `json:"protocol"`
	Description string `json:"description"`
}

// IPRuleUpdate IP规则更新
type IPRuleUpdate struct {
	OldRule IPRuleOperate `json:"oldRule"`
	NewRule IPRuleOperate `json:"newRule"`
}

// ForwardRuleOperate 端口转发规则操作
type ForwardRuleOperate struct {
	ForceDelete bool `json:"forceDelete"`
	Rules       []ForwardRule `json:"rules"`
}

// ForwardRule 端口转发规则
type ForwardRule struct {
	Operation  string `json:"operation" validate:"required,oneof=add remove"`
	Num        string `json:"num"`        // 规则编号
	Protocol   string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
	Interface  string `json:"interface"`  // 网络接口
	Port       string `json:"port" validate:"required"`       // 源端口
	TargetIP   string `json:"targetIP"`                        // 目标IP
	TargetPort string `json:"targetPort" validate:"required"` // 目标端口
}

// InstallRequest 防火墙安装请求
type InstallRequest struct {
	Type string `json:"type" validate:"required,oneof=ufw iptables firewalld"` // 防火墙类型
}

// InstallProgress 安装进度消息
type InstallProgress struct {
	Type     string `json:"type"`     // 消息类型: progress, log, error, complete
	Progress int    `json:"progress"` // 进度百分比 (0-100)
	Message  string `json:"message"`  // 进度消息
	Log      string `json:"log"`      // 日志内容
}