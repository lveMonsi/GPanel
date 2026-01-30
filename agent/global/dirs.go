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