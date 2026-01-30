package firewall

import (
	"gpanel/agent/global"
	"os"
	"path/filepath"
	"strings"
)

const (
	// PingStatusFile Ping状态文件
	PingStatusFile = "ping_status"
)

// LoadPingStatus 加载Ping状态
func LoadPingStatus() string {
	pingFile := filepath.Join(global.GetFirewallDir(), PingStatusFile)

	data, err := os.ReadFile(pingFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "disabled"
		}
		global.Error("read ping status file failed: %v", err)
		return "disabled"
	}

	status := strings.TrimSpace(string(data))
	if status == "1" {
		return "enabled"
	}
	return "disabled"
}

// UpdatePingStatus 更新Ping状态
func UpdatePingStatus(status string) error {
	pingFile := filepath.Join(global.GetFirewallDir(), PingStatusFile)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(pingFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(pingFile, []byte(status), 0644)
}