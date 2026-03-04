package firewall

import (
	"fmt"
	"gpanel/agent/global"
	"os"
	"strings"
)

// LoadPingStatus 加载Ping状态（读取系统实际的ICMP响应设置）
func LoadPingStatus() string {
	// 读取 IPv4 ICMP 设置
	data, err := os.ReadFile("/proc/sys/net/ipv4/icmp_echo_ignore_all")
	if err != nil {
		// 如果无法读取系统文件，返回 disabled
		return "disabled"
	}

	// 检查 IPv6 ICMP 设置（可选）
	v6Data, v6err := os.ReadFile("/proc/sys/net/ipv6/icmp/echo_ignore_all")
	if v6err != nil {
		// IPv6 不存在时，只检查 IPv4
		if strings.TrimSpace(string(data)) == "1" {
			return "enabled"
		}
		return "disabled"
	}

	// IPv4 和 IPv6 都存在时，两者都需要为 1 才算启用
	if strings.TrimSpace(string(data)) == "1" && strings.TrimSpace(string(v6Data)) == "1" {
		return "enabled"
	}
	return "disabled"
}

// UpdatePingStatus 更新Ping状态（真正修改系统ICMP响应设置）
func UpdatePingStatus(status string) error {
	enable := "0"
	if status == "1" {
		enable = "1"
	}

	// 1. 立即生效：写入 /proc/sys/net/ipv4/icmp_echo_ignore_all
	if err := applyPingStatusImmediately(enable); err != nil {
		return err
	}

	// 2. 永久生效：更新 sysctl 配置文件
	if err := updatePingStatusPersistent(enable); err != nil {
		global.Error("failed to update persistent ping status: %v", err)
		// 不返回错误，因为临时设置已经生效
	}

	return nil
}

// applyPingStatusImmediately 立即应用Ping状态（写入proc文件系统）
func applyPingStatusImmediately(enable string) error {
	// IPv4
	cmd := fmt.Sprintf("echo %s > /proc/sys/net/ipv4/icmp_echo_ignore_all", enable)
	if _, err := ExecWithSudo("sh", "-c", cmd); err != nil {
		return fmt.Errorf("failed to apply ipv4 ping status: %v", err)
	}

	// IPv6（如果存在）
	if _, err := os.Stat("/proc/sys/net/ipv6/icmp/echo_ignore_all"); err == nil {
		cmd = fmt.Sprintf("echo %s > /proc/sys/net/ipv6/icmp/echo_ignore_all", enable)
		if _, err := ExecWithSudo("sh", "-c", cmd); err != nil {
			global.Error("failed to apply ipv6 ping status: %v", err)
			// IPv6 失败不阻止整体操作
		}
	}

	return nil
}

// updatePingStatusPersistent 永久保存Ping状态（更新sysctl配置文件）
func updatePingStatusPersistent(enable string) error {
	const confPath = "/etc/sysctl.conf"
	const panelSysctlPath = "/etc/sysctl.d/98-gpanel.conf"

	var targetPath string
	var applyCmd string

	// 确定 sysctl 配置文件路径
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// 如果 /etc/sysctl.conf 不存在，使用专用配置目录
		targetPath = panelSysctlPath
		applyCmd = "sysctl --system"
		// 确保目录存在
		if _, err := ExecWithSudo("mkdir", "-p", "/etc/sysctl.d"); err != nil {
			return fmt.Errorf("failed to create directory /etc/sysctl.d: %v", err)
		}
	} else {
		targetPath = confPath
		applyCmd = "sysctl -p"
	}

	// 读取现有配置文件
	lineBytes, err := os.ReadFile(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %v", targetPath, err)
	}

	// 检查是否存在 IPv6
	hasIpv6 := false
	if _, err := os.Stat("/proc/sys/net/ipv6/icmp/echo_ignore_all"); err == nil {
		hasIpv6 = true
	}

	// 解析并更新配置
	var files []string
	if err == nil {
		files = strings.Split(string(lineBytes), "\n")
	}

	var newFiles []string
	hasIPv4Line, hasIPv6Line := false, false

	for _, line := range files {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "net.ipv4.icmp_echo_ignore_all") {
			newFiles = append(newFiles, "net.ipv4.icmp_echo_ignore_all="+enable)
			hasIPv4Line = true
			continue
		}
		if strings.HasPrefix(trimmedLine, "net.ipv6.icmp.echo_ignore_all") {
			newFiles = append(newFiles, "net.ipv6.icmp.echo_ignore_all="+enable)
			hasIPv6Line = true
			continue
		}
		newFiles = append(newFiles, line)
	}

	// 添加缺失的配置项
	if !hasIPv4Line {
		newFiles = append(newFiles, "net.ipv4.icmp_echo_ignore_all="+enable)
	}
	if hasIpv6 && !hasIPv6Line {
		newFiles = append(newFiles, "net.ipv6.icmp.echo_ignore_all="+enable)
	}

	// 写入配置文件
	content := strings.Join(newFiles, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// 使用临时文件写入，避免权限问题
	tmpFile := "/tmp/gpanel_sysctl.tmp"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// 复制到目标位置
	if _, err := ExecWithSudo("cp", tmpFile, targetPath); err != nil {
		return fmt.Errorf("failed to copy to %s: %v", targetPath, err)
	}

	// 应用配置
	if _, err := ExecWithSudo("sh", "-c", applyCmd); err != nil {
		global.Error("failed to apply sysctl config with '%s': %v", applyCmd, err)
		// 不返回错误，因为临时设置已经生效
	}

	return nil
}
