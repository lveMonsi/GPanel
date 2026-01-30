package firewall

import (
	"gpanel/agent/dto"
)

// FirewallClient 防火墙客户端接口
type FirewallClient interface {
	// 基础操作
	Name() string
	Start() error
	Stop() error
	Restart() error
	Status() (bool, error)
	Version() (string, error)

	// 端口规则
	ListPort() ([]dto.FireInfo, error)
	Port(rule dto.FireInfo, operation string) error

	// IP规则
	ListAddress() ([]dto.FireInfo, error)
	IP(rule dto.FireInfo, operation string) error

	// 端口转发
	ListForward() ([]dto.FireInfo, error)
	AddForward(rule dto.FireInfo) error
	DeleteForward(rule dto.FireInfo) error
	EnableForward() error
}

// NewFirewallClient 创建防火墙客户端
func NewFirewallClient() (FirewallClient, error) {
	// 检测可用的防火墙
	_, err := Exec("which", "ufw")
	if err == nil {
		ufw, err := NewUfwClient()
		if err == nil {
			return ufw, nil
		}
	}

	// 回退到iptables
	iptables, err := NewIptablesClient()
	if err != nil {
		return nil, err
	}
	return iptables, nil
}