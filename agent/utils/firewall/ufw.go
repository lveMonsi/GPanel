package firewall

import (
	"fmt"
	"gpanel/agent/global"
	"gpanel/agent/dto"
	"os/exec"
	"strings"
)

// UfwClient UFW防火墙客户端
type UfwClient struct{}

// NewUfwClient 创建UFW客户端
func NewUfwClient() (*UfwClient, error) {
	return &UfwClient{}, nil
}

// Name 返回防火墙名称
func (u *UfwClient) Name() string {
	return "ufw"
}

// Start 启动防火墙
func (u *UfwClient) Start() error {
	_, err := ExecWithSudo("bash", "-c", "echo y | ufw enable")
	if err != nil {
		return fmt.Errorf("ufw start failed: %w", err)
	}
	return nil
}

// Stop 停止防火墙
func (u *UfwClient) Stop() error {
	_, err := ExecWithSudo("ufw", "--force", "disable")
	if err != nil {
		return fmt.Errorf("ufw stop failed: %w", err)
	}
	return nil
}

// Restart 重启防火墙
func (u *UfwClient) Restart() error {
	if err := u.Stop(); err != nil {
		return err
	}
	return u.Start()
}

// Status 获取防火墙状态
func (u *UfwClient) Status() (bool, error) {
	output, err := ExecWithSudo("ufw", "status")
	if err != nil {
		return false, err
	}
	return strings.Contains(output, "Status: active"), nil
}

// Version 获取防火墙版本
func (u *UfwClient) Version() (string, error) {
	output, err := Exec("ufw", "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// ListPort 列出端口规则
func (u *UfwClient) ListPort() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("ufw", "status", "verbose")
	if err != nil {
		return nil, err
	}

	portInfos := strings.Split(output, "\n")
	var datas []dto.FireInfo
	isStart := false

	for _, line := range portInfos {
		if strings.HasPrefix(line, "-") {
			isStart = true
			continue
		}
		if !isStart {
			continue
		}

		itemFire := u.loadInfo(line, "port")
		if itemFire.Port != "" && itemFire.Port != "Anywhere" && !strings.Contains(itemFire.Port, ".") {
			itemFire.Port = strings.ReplaceAll(itemFire.Port, ":", "-")
			datas = append(datas, itemFire)
		}
	}

	return datas, nil
}

// ListAddress 列出IP规则
func (u *UfwClient) ListAddress() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("ufw", "status", "verbose")
	if err != nil {
		return nil, err
	}

	portInfos := strings.Split(output, "\n")
	var datas []dto.FireInfo
	isStart := false

	for _, line := range portInfos {
		if strings.HasPrefix(line, "-") {
			isStart = true
			continue
		}
		if !isStart {
			continue
		}
		if !strings.Contains(line, " IN") {
			continue
		}

		itemFire := u.loadInfo(line, "address")
		if strings.Contains(itemFire.Port, ".") {
			itemFire.Address += ("-" + itemFire.Port)
			itemFire.Port = ""
		}
		if itemFire.Port == "" && itemFire.Address != "" {
			datas = append(datas, itemFire)
		}
	}

	return datas, nil
}

// ListForward 列出端口转发规则（使用iptables）
func (u *UfwClient) ListForward() ([]dto.FireInfo, error) {
	// UFW本身不支持端口转发，需要使用iptables
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
func (u *UfwClient) Port(rule dto.FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	// 转换策略
	action := "allow"
	if rule.Strategy == "drop" {
		action = "deny"
	}

	// 构建端口字符串
	portSpec := rule.Port
	if rule.Protocol != "" && rule.Protocol != "tcp/udp" {
		portSpec += "/" + rule.Protocol
	}

	// 添加IP地址
	if rule.Address != "" && rule.Address != "0.0.0.0/0" {
		portSpec += " from " + rule.Address
	}

	if operation == "remove" {
		_, err := ExecWithSudo("ufw", "delete", action, portSpec)
		if err != nil {
			return fmt.Errorf("ufw port operation failed: %w", err)
		}
	} else {
		_, err := ExecWithSudo("ufw", action, portSpec)
		if err != nil {
			return fmt.Errorf("ufw port operation failed: %w", err)
		}
	}

	return nil
}

// IP 操作IP规则
func (u *UfwClient) IP(rule dto.FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return fmt.Errorf("invalid operation: %s", operation)
	}

	// 转换策略
	action := "allow"
	if rule.Strategy == "drop" {
		action = "deny"
	}

	if operation == "remove" {
		_, err := ExecWithSudo("ufw", "delete", action, "from", rule.Address)
		if err != nil {
			return fmt.Errorf("ufw ip operation failed: %w", err)
		}
	} else {
		_, err := ExecWithSudo("ufw", action, "from", rule.Address)
		if err != nil {
			return fmt.Errorf("ufw ip operation failed: %w", err)
		}
	}

	return nil
}

// AddForward 添加端口转发规则（使用iptables）
func (u *UfwClient) AddForward(rule dto.FireInfo) error {
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

// DeleteForward 删除端口转发规则（使用iptables）
func (u *UfwClient) DeleteForward(rule dto.FireInfo) error {
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
func (u *UfwClient) EnableForward() error {
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

// loadInfo 解析规则信息
func (u *UfwClient) loadInfo(line string, fireType string) dto.FireInfo {
	fields := strings.Fields(line)
	var itemInfo dto.FireInfo

	if strings.Contains(line, "LIMIT") || strings.Contains(line, "ALLOW FWD") {
		return itemInfo
	}

	if len(fields) < 4 {
		return itemInfo
	}

	if fields[1] == "(v6)" && fireType == "port" {
		return itemInfo
	}

	if fields[0] == "Anywhere" && fireType != "port" {
		itemInfo.Strategy = "drop"
		if fields[1] == "ALLOW" {
			itemInfo.Strategy = "accept"
		}
		if fields[1] == "(v6)" {
			if fields[2] == "ALLOW" {
				itemInfo.Strategy = "accept"
			}
			itemInfo.Address = fields[4]
		} else {
			itemInfo.Address = fields[3]
		}
		return itemInfo
	}

	if strings.Contains(fields[0], "/") {
		itemInfo.Port = strings.Split(fields[0], "/")[0]
		itemInfo.Protocol = strings.Split(fields[0], "/")[1]
	} else {
		itemInfo.Port = fields[0]
		itemInfo.Protocol = "tcp/udp"
	}

	itemInfo.Family = "ipv4"
	if fields[1] == "ALLOW" {
		itemInfo.Strategy = "accept"
	} else {
		itemInfo.Strategy = "drop"
	}

	itemInfo.Address = fields[3]

	return itemInfo
}