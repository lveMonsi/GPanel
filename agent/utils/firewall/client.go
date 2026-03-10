package firewall

import (
	"errors"
	"gpanel/agent/dto"
)

// ErrFirewallNotInstalled 防火墙未安装错误
var ErrFirewallNotInstalled = errors.New("firewall not installed")

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

// FirewallType 防火墙类型
type FirewallType string

const (
	FirewallTypeUfw        FirewallType = "ufw"
	FirewallTypeFirewalld  FirewallType = "firewalld"
	FirewallTypeIptables   FirewallType = "iptables"
	FirewallTypeNone       FirewallType = "none"
)

// CheckFirewallInstalled 检测系统安装的防火墙类型
// 返回值: 防火墙类型，是否安装
func CheckFirewallInstalled() (FirewallType, bool) {
	// 优先检测 ufw
	_, err := Exec("which", "ufw")
	if err == nil {
		return FirewallTypeUfw, true
	}

	// 检测 firewalld
	_, err = Exec("which", "firewall-cmd")
	if err == nil {
		return FirewallTypeFirewalld, true
	}

	// 检测 iptables 命令是否存在
	_, err = Exec("which", "iptables")
	if err == nil {
		return FirewallTypeIptables, true
	}

	return FirewallTypeNone, false
}

// NewFirewallClient 创建防火墙客户端
func NewFirewallClient() (FirewallClient, error) {
	// 检测可用的防火墙
	firewallType, installed := CheckFirewallInstalled()
	if !installed {
		return nil, ErrFirewallNotInstalled
	}

	switch firewallType {
	case FirewallTypeUfw:
		ufw, err := NewUfwClient()
		if err == nil {
			return ufw, nil
		}
		// ufw 创建失败，回退到 iptables
	case FirewallTypeFirewalld:
		// firewalld 暂不支持，回退到 iptables
	}

	// 回退到 iptables
	iptables, err := NewIptablesClient()
	if err != nil {
		return nil, err
	}
	return iptables, nil
}