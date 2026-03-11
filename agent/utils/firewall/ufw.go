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
	// 解析版本号，格式: "ufw 0.36.2 Copyright 2008-2023 Canonical Ltd."
	fields := strings.Fields(output)
	if len(fields) >= 2 {
		return fields[1], nil
	}
	return strings.TrimSpace(output), nil
}

// ListPort 列出端口规则
func (u *UfwClient) ListPort() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("ufw", "status", "verbose")
	if err != nil {
		return nil, err
	}

	// 检查防火墙是否禁用
	if strings.Contains(output, "Status: inactive") {
		// 防火墙禁用时，使用 show added 获取已添加的规则
		return u.listPortFromShowAdded()
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

// listPortFromShowAdded 从 ufw show added 获取端口规则（防火墙禁用时使用）
func (u *UfwClient) listPortFromShowAdded() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("ufw", "show", "added")
	if err != nil {
		return nil, err
	}

	var datas []dto.FireInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Added") || strings.HasPrefix(line, "See") {
			continue
		}

		itemFire := u.parseAddedRule(line, "port")
		if itemFire.Port != "" && itemFire.Port != "Anywhere" && !strings.Contains(itemFire.Port, ".") {
			itemFire.Port = strings.ReplaceAll(itemFire.Port, ":", "-")
			datas = append(datas, itemFire)
		}
	}

	return datas, nil
}

// parseAddedRule 解析 ufw show added 输出的规则
func (u *UfwClient) parseAddedRule(line string, fireType string) dto.FireInfo {
	var itemInfo dto.FireInfo

	// 移除行首的空格和制表符
	line = strings.TrimSpace(line)
	if line == "" {
		return itemInfo
	}

	// 跳过 (v6) 规则（端口规则）
	if strings.Contains(line, "(v6)") && fireType == "port" {
		return itemInfo
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return itemInfo
	}

	// 检测规则类型
	// 格式1: "Anywhere ALLOW 192.168.1.1" - 纯 IP 规则
	// 格式2: "allow from 192.168.1.1" - 纯 IP 规则（无端口）
	// 格式3: "80/tcp" 或 "allow 80/tcp" - 纯端口规则
	// 格式4: "allow from 192.168.1.1 to any port 22 proto tcp" - 带 IP 的端口规则

	// 检查是否包含 port 关键字（区分纯 IP 规则和带 IP 的端口规则）
	hasPort := false
	for _, field := range fields {
		if field == "port" {
			hasPort = true
			break
		}
	}

	// 检查是否是纯 IP 规则（没有端口号）
	// 纯 IP 规则: 包含 "from" 但不包含 "port"
	isIPRule := false
	for _, field := range fields {
		if field == "from" {
			isIPRule = true
			break
		}
	}

	// 格式1: "Anywhere ALLOW 192.168.1.1" - 这是纯 IP 规则
	if fields[0] == "Anywhere" {
		isIPRule = true
	}

	// 如果包含 port 关键字，则不是纯 IP 规则，而是带 IP 的端口规则
	if hasPort {
		isIPRule = false
	}

	// 如果是端口规则请求，但检测到是纯 IP 规则，返回空结构体
	if fireType == "port" && isIPRule {
		return itemInfo
	}

	// 如果是 IP 规则请求，但检测到不是 IP 规则，返回空结构体
	if fireType == "address" && !isIPRule {
		return itemInfo
	}

	// 解析纯 IP 规则
	if isIPRule {
		// 格式1: "Anywhere ALLOW 192.168.1.1"
		if fields[0] == "Anywhere" && len(fields) >= 3 {
			if fields[1] == "ALLOW" {
				itemInfo.Strategy = "accept"
			} else if fields[1] == "DENY" {
				itemInfo.Strategy = "drop"
			}
			itemInfo.Address = fields[2]
			return itemInfo
		}

		// 格式2: "allow from 192.168.1.1" 或 "deny from 192.168.1.1"
		for i, field := range fields {
			if field == "allow" {
				itemInfo.Strategy = "accept"
			} else if field == "deny" {
				itemInfo.Strategy = "drop"
			} else if field == "from" && i+1 < len(fields) {
				itemInfo.Address = fields[i+1]
			}
		}
		return itemInfo
	}

	// 解析端口规则（包括带 IP 的端口规则）
	// 格式: "80/tcp" 或 "allow 80/tcp" 或 "allow from 192.168.1.1 to any port 22 proto tcp"
	var portSpec string
	action := "allow" // 默认允许

	// 提取策略和端口
	for i, field := range fields {
		if field == "allow" || field == "deny" {
			action = field
		} else if field == "port" && i+1 < len(fields) {
			// 带 IP 的端口规则: "port 22" 后面跟着端口号
			portSpec = fields[i+1]
		}
	}

	// 如果没有找到 port 关键字，尝试直接解析端口规格
	if portSpec == "" {
		for i, field := range fields {
			if field == "allow" || field == "deny" {
				if i+1 < len(fields) {
					portSpec = fields[i+1]
				}
				break
			}
		}
	}

	// 如果没有找到 action 关键字，最后一个字段就是端口规格
	if portSpec == "" && len(fields) > 0 {
		portSpec = fields[len(fields)-1]
	}

	// 解析端口和协议
	if strings.Contains(portSpec, "/") {
		parts := strings.Split(portSpec, "/")
		itemInfo.Port = parts[0]
		if len(parts) > 1 {
			itemInfo.Protocol = parts[1]
		}
	} else if portSpec != "" {
		itemInfo.Port = portSpec
		itemInfo.Protocol = "tcp/udp"
	}

	// 设置策略
	if action == "allow" {
		itemInfo.Strategy = "accept"
	} else {
		itemInfo.Strategy = "drop"
	}

	// 提取 IP 地址（如果有）
	for i, field := range fields {
		if field == "from" && i+1 < len(fields) {
			itemInfo.Address = fields[i+1]
			break
		}
	}

	return itemInfo
}

// ListAddress 列出IP规则
func (u *UfwClient) ListAddress() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("ufw", "status", "verbose")
	if err != nil {
		return nil, err
	}

	// 检查防火墙是否禁用
	if strings.Contains(output, "Status: inactive") {
		// 防火墙禁用时，使用 show added 获取已添加的规则
		return u.listAddressFromShowAdded()
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

// listAddressFromShowAdded 从 ufw show added 获取 IP 规则（防火墙禁用时使用）
func (u *UfwClient) listAddressFromShowAdded() ([]dto.FireInfo, error) {
	output, err := ExecWithSudo("ufw", "show", "added")
	if err != nil {
		return nil, err
	}

	var datas []dto.FireInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Added") || strings.HasPrefix(line, "See") {
			continue
		}

		// 解析 IP 规则，格式如: "Anywhere ALLOW 192.168.1.1" 或 "allow from 192.168.1.1"
		itemFire := u.parseAddedIPRule(line)
		if itemFire.Address != "" && itemFire.Port == "" {
			datas = append(datas, itemFire)
		}
	}

	return datas, nil
}

// parseAddedIPRule 解析 ufw show added 输出的 IP 规则
func (u *UfwClient) parseAddedIPRule(line string) dto.FireInfo {
	var itemInfo dto.FireInfo

	line = strings.TrimSpace(line)
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return itemInfo
	}

	// 格式1: "Anywhere ALLOW 192.168.1.1"
	// 格式2: "allow from 192.168.1.1"
	// 格式3: "deny from 192.168.1.1"

	// 检查格式1
	if fields[0] == "Anywhere" && len(fields) >= 3 {
		if fields[1] == "ALLOW" {
			itemInfo.Strategy = "accept"
		} else {
			itemInfo.Strategy = "drop"
		}
		itemInfo.Address = fields[2]
		return itemInfo
	}

	// 检查格式2和3
	for i, field := range fields {
		if field == "allow" {
			itemInfo.Strategy = "accept"
		} else if field == "deny" {
			itemInfo.Strategy = "drop"
		} else if field == "from" && i+1 < len(fields) {
			itemInfo.Address = fields[i+1]
			break
		}
	}

	return itemInfo
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

	// 当协议为 tcp/udp 时，需要分别对 tcp 和 udp 执行操作
	protocols := []string{rule.Protocol}
	if rule.Protocol == "" || rule.Protocol == "tcp/udp" {
		protocols = []string{"tcp", "udp"}
	}

	var lastErr error
	for _, proto := range protocols {
		var err error
		if operation == "remove" {
			if rule.Address != "" && rule.Address != "0.0.0.0/0" {
				// 删除带 IP 的规则: ufw delete allow from <IP> to any port <PORT> proto <PROTO>
				_, err = ExecWithSudo("ufw", "delete", action, "from", rule.Address, "to", "any", "port", rule.Port, "proto", proto)
			} else {
				// 删除不带 IP 的规则: ufw delete allow <PORT>/<PROTO>
				portSpec := rule.Port + "/" + proto
				_, err = ExecWithSudo("ufw", "delete", action, portSpec)
			}
		} else {
			if rule.Address != "" && rule.Address != "0.0.0.0/0" {
				// 添加带 IP 的规则: ufw allow from <IP> to any port <PORT> proto <PROTO>
				_, err = ExecWithSudo("ufw", action, "from", rule.Address, "to", "any", "port", rule.Port, "proto", proto)
			} else {
				// 添加不带 IP 的规则: ufw allow <PORT>/<PROTO>
				portSpec := rule.Port + "/" + proto
				_, err = ExecWithSudo("ufw", action, portSpec)
			}
		}
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("ufw port operation failed: %w", lastErr)
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
// UFW status verbose 输出格式示例：
// IPv4: "22/tcp                   ALLOW IN    192.168.1.1"
// IPv4: "80/tcp                   ALLOW IN    Anywhere"
// IPv6: "22/tcp (v6)              ALLOW IN    Anywhere (v6)"
// 纯IP: "Anywhere                ALLOW IN    192.168.1.1"
func (u *UfwClient) loadInfo(line string, fireType string) dto.FireInfo {
	fields := strings.Fields(line)
	var itemInfo dto.FireInfo

	if strings.Contains(line, "LIMIT") || strings.Contains(line, "ALLOW FWD") {
		return itemInfo
	}

	if len(fields) < 4 {
		return itemInfo
	}

	// 检测是否是 IPv6 规则
	isIPv6 := false
	for _, field := range fields {
		if field == "(v6)" {
			isIPv6 = true
			break
		}
	}

	// 跳过 IPv6 端口规则（端口规则只显示 IPv4）
	if isIPv6 && fireType == "port" {
		return itemInfo
	}

	// 解析纯 IP 规则（Anywhere 开头的规则）
	if fields[0] == "Anywhere" && fireType != "port" {
		itemInfo.Strategy = "drop"
		// 查找 ALLOW 或 DENY
		for i, field := range fields {
			if field == "ALLOW" {
				itemInfo.Strategy = "accept"
				// ALLOW 后面可能跟着 IN/OUT，IP 地址在后面
				// 格式: "Anywhere ALLOW IN 192.168.1.1"
				if i+2 < len(fields) {
					itemInfo.Address = fields[i+2]
				} else if i+1 < len(fields) {
					itemInfo.Address = fields[i+1]
				}
				break
			} else if field == "DENY" {
				itemInfo.Strategy = "drop"
				if i+2 < len(fields) {
					itemInfo.Address = fields[i+2]
				} else if i+1 < len(fields) {
					itemInfo.Address = fields[i+1]
				}
				break
			}
		}
		// 过滤掉 Anywhere (v6) 这样的地址
		if itemInfo.Address == "Anywhere" || strings.Contains(itemInfo.Address, "(v6)") {
			itemInfo.Address = ""
		}
		return itemInfo
	}

	// 解析端口规则
	// 格式: "22/tcp ALLOW IN 192.168.1.1" 或 "22/tcp ALLOW 192.168.1.1"
	if strings.Contains(fields[0], "/") {
		itemInfo.Port = strings.Split(fields[0], "/")[0]
		itemInfo.Protocol = strings.Split(fields[0], "/")[1]
	} else {
		itemInfo.Port = fields[0]
		itemInfo.Protocol = "tcp/udp"
	}

	// 查找策略和 IP 地址
	// 遍历字段找到 ALLOW/DENY，然后提取后面的 IP 地址
	itemInfo.Family = "ipv4"
	if isIPv6 {
		itemInfo.Family = "ipv6"
	}

	for i, field := range fields {
		if field == "ALLOW" {
			itemInfo.Strategy = "accept"
			// ALLOW 后面可能跟着 IN/OUT，IP 地址可能在 i+2 或 i+1 位置
			if i+2 < len(fields) && (fields[i+1] == "IN" || fields[i+1] == "OUT") {
				itemInfo.Address = fields[i+2]
			} else if i+1 < len(fields) {
				itemInfo.Address = fields[i+1]
			}
			break
		} else if field == "DENY" {
			itemInfo.Strategy = "drop"
			if i+2 < len(fields) && (fields[i+1] == "IN" || fields[i+1] == "OUT") {
				itemInfo.Address = fields[i+2]
			} else if i+1 < len(fields) {
				itemInfo.Address = fields[i+1]
			}
			break
		}
	}

	// 过滤掉 Anywhere 这样的通用地址
	if itemInfo.Address == "Anywhere" || itemInfo.Address == "Anywhere" || strings.Contains(itemInfo.Address, "(v6)") {
		itemInfo.Address = ""
	}

	return itemInfo
}