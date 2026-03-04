package firewall

import (
	"fmt"
	"gpanel/agent/global"
	"gpanel/agent/dto"
	"os/exec"
	"strings"
)

// IptablesClient iptables防火墙客户端
type IptablesClient struct{}

// NewIptablesClient 创建iptables客户端
func NewIptablesClient() (*IptablesClient, error) {
	return &IptablesClient{}, nil
}

// Name 返回防火墙名称
func (i *IptablesClient) Name() string {
	return "iptables"
}

// Start 启动防火墙
func (i *IptablesClient) Start() error {
	return nil // iptables默认就是启动的
}

// Stop 停止防火墙
func (i *IptablesClient) Stop() error {
	// 清空所有规则
	_, err := ExecWithSudo("iptables", "-F")
	if err != nil {
		return fmt.Errorf("iptables stop failed: %w", err)
	}
	return nil
}

// Restart 重启防火墙
func (i *IptablesClient) Restart() error {
	return nil
}

// Status 获取防火墙状态
func (i *IptablesClient) Status() (bool, error) {
	// 检查是否有规则
	output, err := ExecWithSudo("iptables", "-L", "-n")
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(output)) > 0, nil
}

// Version 获取防火墙版本
func (i *IptablesClient) Version() (string, error) {
	output, err := Exec("iptables", "--version")
	if err != nil {
		return "", err
	}
	// 解析版本号，格式: "iptables v1.8.7 (nf_tables)"
	fields := strings.Fields(output)
	if len(fields) >= 2 {
		// 移除 v 前缀
		version := strings.TrimPrefix(fields[1], "v")
		return version, nil
	}
	return strings.TrimSpace(output), nil
}

// ListPort 列出端口规则
func (i *IptablesClient) ListPort() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("iptables", "-L", "INPUT", "-n", "--line-numbers")
	if err != nil {
		return nil, err
	}

	var rules []dto.FireInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chain") || strings.HasPrefix(line, "num") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// 解析规则格式
		rule := dto.FireInfo{
			Num:      fields[0],
			Strategy: fields[2],
			Protocol: fields[1],
		}

		// 转换策略
		if rule.Strategy == "ACCEPT" {
			rule.Strategy = "accept"
		} else if rule.Strategy == "DROP" {
			rule.Strategy = "drop"
		}

		// 提取端口
		for idx := 3; idx < len(fields); idx++ {
			if strings.HasPrefix(fields[idx], "dpt:") {
				rule.Port = strings.TrimPrefix(fields[idx], "dpt:")
				break
			}
		}

		// 提取源地址
		for idx := 3; idx < len(fields); idx++ {
			if fields[idx] != "0.0.0.0/0" && !strings.Contains(fields[idx], "dpt:") {
				rule.Address = fields[idx]
				break
			}
		}

		if rule.Port != "" {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// ListAddress 列出IP规则
func (i *IptablesClient) ListAddress() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("iptables", "-L", "INPUT", "-n", "--line-numbers")
	if err != nil {
		return nil, err
	}

	var rules []dto.FireInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chain") || strings.HasPrefix(line, "num") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		rule := dto.FireInfo{
			Num:      fields[0],
			Strategy: fields[2],
			Protocol: fields[1],
		}

		if rule.Strategy == "ACCEPT" {
			rule.Strategy = "accept"
		} else if rule.Strategy == "DROP" {
			rule.Strategy = "drop"
		}

		// 提取源地址（排除端口规则）
		for idx := 3; idx < len(fields); idx++ {
			if fields[idx] != "0.0.0.0/0" && !strings.Contains(fields[idx], "dpt:") {
				rule.Address = fields[idx]
				break
			}
		}

		// 只返回纯IP规则
		if rule.Address != "" && !strings.Contains(line, "dpt:") {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// ListForward 列出端口转发规则
func (i *IptablesClient) ListForward() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("iptables", "-t", "nat", "-L", "PREROUTING", "-n", "--line-numbers")
	if err != nil {
		return nil, err
	}

	var rules []dto.FireInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chain") || strings.HasPrefix(line, "num") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		rule := dto.FireInfo{
			Num:      fields[0],
			Protocol: fields[3],
		}

		// 提取源端口
		for idx := 0; idx < len(fields); idx++ {
			if strings.Contains(fields[idx], "dpt:") {
				rule.Port = strings.TrimPrefix(fields[idx], "dpt:")
				break
			}
		}

		// 提取目标IP和端口
		for idx := 0; idx < len(fields); idx++ {
			if strings.Contains(fields[idx], "to:") {
				toStr := strings.TrimPrefix(fields[idx], "to:")
				parts := strings.Split(toStr, ":")
				if len(parts) >= 2 {
					rule.TargetIP = parts[0]
					rule.TargetPort = parts[1]
				}
				break
			}
		}

		if rule.Port != "" {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// Port 操作端口规则
func (i *IptablesClient) Port(rule dto.FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	action := "ACCEPT"
	if rule.Strategy == "drop" {
		action = "DROP"
	}

	// 构建规则
	if rule.Address != "" && rule.Address != "0.0.0.0/0" {
		// 带源地址的规则
		if operation == "add" {
			_, err := ExecWithSudo("iptables", "-A", "INPUT", "-p", rule.Protocol, "-s", rule.Address, "--dport", rule.Port, "-j", action)
			return err
		} else {
			_, err := ExecWithSudo("iptables", "-D", "INPUT", "-p", rule.Protocol, "-s", rule.Address, "--dport", rule.Port, "-j", action)
			return err
		}
	} else {
		// 不带源地址的规则
		if operation == "add" {
			_, err := ExecWithSudo("iptables", "-A", "INPUT", "-p", rule.Protocol, "--dport", rule.Port, "-j", action)
			return err
		} else {
			_, err := ExecWithSudo("iptables", "-D", "INPUT", "-p", rule.Protocol, "--dport", rule.Port, "-j", action)
			return err
		}
	}
}

// IP 操作IP规则
func (i *IptablesClient) IP(rule dto.FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	action := "ACCEPT"
	if rule.Strategy == "drop" {
		action = "DROP"
	}

	if rule.Protocol != "" && rule.Protocol != "tcp/udp" {
		// 带协议的规则
		if operation == "add" {
			_, err := ExecWithSudo("iptables", "-A", "INPUT", "-p", rule.Protocol, "-s", rule.Address, "-j", action)
			return err
		} else {
			_, err := ExecWithSudo("iptables", "-D", "INPUT", "-p", rule.Protocol, "-s", rule.Address, "-j", action)
			return err
		}
	} else {
		// 不带协议的规则
		if operation == "add" {
			_, err := ExecWithSudo("iptables", "-A", "INPUT", "-s", rule.Address, "-j", action)
			return err
		} else {
			_, err := ExecWithSudo("iptables", "-D", "INPUT", "-s", rule.Address, "-j", action)
			return err
		}
	}
}

// AddForward 添加端口转发规则
func (i *IptablesClient) AddForward(rule dto.FireInfo) error {
	if rule.TargetIP == "" || rule.TargetIP == "127.0.0.1" || rule.TargetIP == "localhost" {
		// 转发到本地（REDIRECT）
		if rule.Interface != "" {
			_, err := ExecWithSudo("iptables", "-t", "nat", "-A", "PREROUTING", "-i", rule.Interface, "-p", rule.Protocol, "--dport", rule.Port, "-j", "REDIRECT", "--to-port", rule.TargetPort)
			if err != nil {
				return fmt.Errorf("add forward failed: %w", err)
			}
		} else {
			_, err := ExecWithSudo("iptables", "-t", "nat", "-A", "PREROUTING", "-p", rule.Protocol, "--dport", rule.Port, "-j", "REDIRECT", "--to-port", rule.TargetPort)
			if err != nil {
				return fmt.Errorf("add forward failed: %w", err)
			}
		}
	} else {
		// 转发到远程（DNAT）
		if rule.Interface != "" {
			_, err := ExecWithSudo("iptables", "-t", "nat", "-A", "PREROUTING", "-i", rule.Interface, "-p", rule.Protocol, "--dport", rule.Port, "-j", "DNAT", "--to-destination", rule.TargetIP+":"+rule.TargetPort)
			if err != nil {
				return fmt.Errorf("add forward failed: %w", err)
			}
		} else {
			_, err := ExecWithSudo("iptables", "-t", "nat", "-A", "PREROUTING", "-p", rule.Protocol, "--dport", rule.Port, "-j", "DNAT", "--to-destination", rule.TargetIP+":"+rule.TargetPort)
			if err != nil {
				return fmt.Errorf("add forward failed: %w", err)
			}
		}

		// MASQUERADE
		_, err := ExecWithSudo("iptables", "-t", "nat", "-A", "POSTROUTING", "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", rule.TargetPort, "-j", "MASQUERADE")
		if err != nil {
			return fmt.Errorf("add masquerade failed: %w", err)
		}

		// FORWARD ACCEPT
		_, err = ExecWithSudo("iptables", "-A", "FORWARD", "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", rule.TargetPort, "-j", "ACCEPT")
		if err != nil {
			return fmt.Errorf("add forward accept failed: %w", err)
		}
		_, err = ExecWithSudo("iptables", "-A", "FORWARD", "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", rule.TargetPort, "-j", "ACCEPT")
		if err != nil {
			return fmt.Errorf("add forward accept failed: %w", err)
		}
	}

	return nil
}

// DeleteForward 删除端口转发规则
func (i *IptablesClient) DeleteForward(rule dto.FireInfo) error {
	// 使用规则编号删除
	if rule.Num != "" {
		_, err := ExecWithSudo("iptables", "-t", "nat", "-D", "PREROUTING", rule.Num)
		if err != nil {
			return fmt.Errorf("delete forward failed: %w", err)
		}

		if rule.TargetIP != "" && rule.TargetIP != "127.0.0.1" && rule.TargetIP != "localhost" {
			// 清理相关规则
			ExecWithSudo("iptables", "-t", "nat", "-D", "POSTROUTING", "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", rule.TargetPort, "-j", "MASQUERADE")
			ExecWithSudo("iptables", "-D", "FORWARD", "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", rule.TargetPort, "-j", "ACCEPT")
			ExecWithSudo("iptables", "-D", "FORWARD", "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", rule.TargetPort, "-j", "ACCEPT")
		}
	}

	return nil
}

// EnableForward 启用端口转发
func (i *IptablesClient) EnableForward() error {
	// 启用IP转发
	cmd := exec.Command("bash", "-c", "echo 1 > /proc/sys/net/ipv4/ip_forward")
	err := cmd.Run()
	if err != nil {
		global.Error("Failed to enable IP forward: " + err.Error())
		return err
	}
	global.Info("IP forward enabled")
	return nil
}