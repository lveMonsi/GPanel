package global

import (
	"path/filepath"
)

// GetFirewallDir 获取防火墙规则目录
func GetFirewallDir() string {
	return filepath.Join(DataDir, "firewall")
}

// GetTempDir 获取临时文件目录
func GetTempDir() string {
	return filepath.Join(DataDir, "temp")
}

// GetSSHDir 获取 SSH 数据目录
func GetSSHDir() string {
	return filepath.Join(DataDir, "ssh")
}

// GetKnownHostsFile 获取已知主机文件路径
func GetKnownHostsFile() string {
	return filepath.Join(GetSSHDir(), "known_hosts")
}
