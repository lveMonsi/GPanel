package service

import (
	"fmt"
	"gpanel/agent/dto"
	"gpanel/agent/global"
	"gpanel/agent/utils/firewall"
	"strings"
	"sync"
)

// FirewallService 防火墙服务
type FirewallService struct {
	client firewall.FirewallClient
}

// NewFirewallService 创建防火墙服务
func NewFirewallService() *FirewallService {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		global.Error("Failed to create firewall client: " + err.Error())
		return &FirewallService{}
	}
	return &FirewallService{client: client}
}

// LoadBaseInfo 加载防火墙基础信息
func (s *FirewallService) LoadBaseInfo() (dto.FirewallBaseInfo, error) {
	// 检测防火墙是否安装
	firewallType, installed := firewall.CheckFirewallInstalled()
	if !installed {
		return dto.FirewallBaseInfo{
			Name:    "未安装",
			IsExist: false,
		}, nil
	}

	if s.client == nil {
		return dto.FirewallBaseInfo{
			Name:    string(firewallType),
			IsExist: false,
		}, nil
	}

	// 验证防火墙客户端是否真正可用
	// 尝试获取状态，如果失败则认为防火墙不可用
	_, err := s.client.Status()
	if err != nil {
		global.Error("Firewall client status check failed: " + err.Error())
		return dto.FirewallBaseInfo{
			Name:    string(firewallType),
			IsExist: false,
		}, nil
	}

	var baseInfo dto.FirewallBaseInfo
	baseInfo.Version = "-"
	baseInfo.Name = s.client.Name()
	baseInfo.IsExist = true

	var wg sync.WaitGroup
	wg.Add(2)

	// 并发加载状态
	go func() {
		defer wg.Done()
		baseInfo.PingStatus = firewall.LoadPingStatus()
		version, _ := s.client.Version()
		if version != "" {
			baseInfo.Version = version
		}
	}()

	go func() {
		defer wg.Done()
		isActive, _ := s.client.Status()
		baseInfo.IsActive = isActive
	}()

	wg.Wait()
	return baseInfo, nil
}

// SearchRules 搜索防火墙规则
func (s *FirewallService) SearchRules(req dto.RuleSearch) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("firewall client not available")
	}

	var rules []dto.FireInfo
	var err error

	switch req.Type {
	case "port":
		rules, err = s.client.ListPort()
	case "forward":
		rules, err = s.client.ListForward()
	case "address":
		rules, err = s.client.ListAddress()
	default:
		return nil, fmt.Errorf("invalid type: %s", req.Type)
	}

	if err != nil {
		return nil, err
	}

	// 端口规则时检查端口占用状态
	if req.Type == "port" {
		s.fillPortUsedStatus(rules)
	}

	// 信息过滤
	if req.Info != "" {
		var filtered []dto.FireInfo
		for _, rule := range rules {
			if contains(rule.Port, req.Info) || contains(rule.Address, req.Info) || contains(rule.TargetPort, req.Info) || contains(rule.TargetIP, req.Info) {
				filtered = append(filtered, rule)
			}
		}
		rules = filtered
	}

	// 策略过滤
	if req.Strategy != "" {
		var filtered []dto.FireInfo
		for _, rule := range rules {
			if rule.Strategy == req.Strategy {
				filtered = append(filtered, rule)
			}
		}
		rules = filtered
	}

	// 分页
	total := len(rules)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return map[string]interface{}{
		"total": total,
		"items": rules[start:end],
	}, nil
}

// fillPortUsedStatus 填充端口占用状态
func (s *FirewallService) fillPortUsedStatus(rules []dto.FireInfo) {
	// 收集所有端口
	var ports []string
	for _, rule := range rules {
		if rule.Port != "" {
			// 解析端口范围
			parsedPorts := firewall.ParsePortRange(rule.Port)
			ports = append(ports, parsedPorts...)
		}
	}

	// 批量检查端口状态
	if len(ports) > 0 {
		portStatusMap := firewall.CheckPortsUsed(ports, "")

		// 填充状态
		for i := range rules {
			if rules[i].Port != "" {
				parsedPorts := firewall.ParsePortRange(rules[i].Port)
				if len(parsedPorts) > 0 {
					// 检查第一个端口的状态
					protocol := rules[i].Protocol
					if protocol == "" || protocol == "tcp/udp" {
						protocol = "tcp"
					}
					key := parsedPorts[0] + "_" + protocol
					if status, exists := portStatusMap[key]; exists {
						if status.Used {
							rules[i].UsedStatus = "used"
						} else {
							rules[i].UsedStatus = "unused"
						}
					} else {
						rules[i].UsedStatus = "unused"
					}
				}
			}
		}
	}
}

// OperateFirewall 操作防火墙
func (s *FirewallService) OperateFirewall(req dto.FirewallOperation) error {
	if s.client == nil {
		return fmt.Errorf("firewall client not available")
	}

	switch req.Operation {
	case "start":
		return s.client.Start()
	case "stop":
		return s.client.Stop()
	case "restart":
		return s.client.Restart()
	case "enableBanPing":
		return firewall.UpdatePingStatus("1")
	case "disableBanPing":
		return firewall.UpdatePingStatus("0")
	default:
		return fmt.Errorf("invalid operation: %s", req.Operation)
	}
}

// OperatePortRule 操作端口规则
func (s *FirewallService) OperatePortRule(req dto.PortRuleOperate) error {
	if s.client == nil {
		return fmt.Errorf("firewall client not available")
	}

	rule := dto.FireInfo{
		Port:     req.Port,
		Protocol: req.Protocol,
		Strategy: req.Strategy,
		Address:  req.Address,
	}

	return s.client.Port(rule, req.Operation)
}

// UpdatePortRule 更新端口规则
func (s *FirewallService) UpdatePortRule(req dto.PortRuleUpdate) error {
	if s.client == nil {
		return fmt.Errorf("firewall client not available")
	}

	// 先删除旧规则
	oldRule := dto.FireInfo{
		Port:     req.OldRule.Port,
		Protocol: req.OldRule.Protocol,
		Strategy: req.OldRule.Strategy,
		Address:  req.OldRule.Address,
	}
	if err := s.client.Port(oldRule, "remove"); err != nil {
		return fmt.Errorf("failed to remove old rule: %w", err)
	}

	// 添加新规则
	newRule := dto.FireInfo{
		Port:     req.NewRule.Port,
		Protocol: req.NewRule.Protocol,
		Strategy: req.NewRule.Strategy,
		Address:  req.NewRule.Address,
	}
	if err := s.client.Port(newRule, "add"); err != nil {
		return fmt.Errorf("failed to add new rule: %w", err)
	}

	return nil
}

// OperateIPRule 操作IP规则
func (s *FirewallService) OperateIPRule(req dto.IPRuleOperate) error {
	if s.client == nil {
		return fmt.Errorf("firewall client not available")
	}

	rule := dto.FireInfo{
		Address:  req.Address,
		Strategy: req.Strategy,
		Protocol: req.Protocol,
	}

	return s.client.IP(rule, req.Operation)
}

// UpdateIPRule 更新IP规则
func (s *FirewallService) UpdateIPRule(req dto.IPRuleUpdate) error {
	if s.client == nil {
		return fmt.Errorf("firewall client not available")
	}

	// 先删除旧规则
	oldRule := dto.FireInfo{
		Address:  req.OldRule.Address,
		Strategy: req.OldRule.Strategy,
		Protocol: req.OldRule.Protocol,
	}
	if err := s.client.IP(oldRule, "remove"); err != nil {
		return fmt.Errorf("failed to remove old rule: %w", err)
	}

	// 添加新规则
	newRule := dto.FireInfo{
		Address:  req.NewRule.Address,
		Strategy: req.NewRule.Strategy,
		Protocol: req.NewRule.Protocol,
	}
	if err := s.client.IP(newRule, "add"); err != nil {
		return fmt.Errorf("failed to add new rule: %w", err)
	}

	return nil
}

// OperateForwardRule 操作端口转发规则
func (s *FirewallService) OperateForwardRule(req dto.ForwardRuleOperate) error {
	if s.client == nil {
		return fmt.Errorf("firewall client not available")
	}

	for _, rule := range req.Rules {
		firewallRule := dto.FireInfo{
			Protocol:   rule.Protocol,
			Port:       rule.Port,
			TargetIP:   rule.TargetIP,
			TargetPort: rule.TargetPort,
			Num:        rule.Num,
			Interface:  rule.Interface,
		}

		switch rule.Operation {
		case "add":
			if err := s.client.AddForward(firewallRule); err != nil {
				return fmt.Errorf("failed to add forward rule: %w", err)
			}
		case "remove":
			if err := s.client.DeleteForward(firewallRule); err != nil {
				if req.ForceDelete {
					global.Error("Failed to delete forward rule: " + err.Error())
					continue
				}
				return fmt.Errorf("failed to delete forward rule: %w", err)
			}
		default:
			return fmt.Errorf("invalid operation: %s", rule.Operation)
		}
	}

	return nil
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// InstallFirewall 安装防火墙
func (s *FirewallService) InstallFirewall(firewallType string, progressChan chan<- dto.InstallProgress) {
	defer close(progressChan)

	sendProgress := func(progressType string, progress int, message, logContent string) {
		progressChan <- dto.InstallProgress{
			Type:     progressType,
			Progress: progress,
			Message:  message,
			Log:      logContent,
		}
	}

	sendLog := func(logContent string) {
		sendProgress("log", 0, "", logContent)
	}

	// 检测系统包管理器
	var pkgManager string
	var installCmd []string
	var checkCmd string
	var serviceName string

	switch firewallType {
	case "ufw":
		sendProgress("progress", 5, "准备安装 UFW 防火墙...", "")
		sendLog("[INFO] UFW (Uncomplicated Firewall) 是 Ubuntu/Debian 的默认防火墙工具")
		pkgManager = detectPackageManager(sendLog)
		if pkgManager == "" {
			sendProgress("error", 0, "不支持的系统，无法自动安装", "")
			return
		}
		switch pkgManager {
		case "apt":
			installCmd = []string{"apt", "install", "-y", "ufw"}
		case "yum":
			installCmd = []string{"yum", "install", "-y", "ufw"}
		case "dnf":
			installCmd = []string{"dnf", "install", "-y", "ufw"}
		default:
			sendProgress("error", 0, "无法确定包管理器", "")
			return
		}
		checkCmd = "ufw"
		serviceName = "ufw"

	case "iptables":
		sendProgress("progress", 5, "准备安装 iptables 防火墙...", "")
		sendLog("[INFO] iptables 是 Linux 内核级别的防火墙工具")
		pkgManager = detectPackageManager(sendLog)
		if pkgManager == "" {
			sendProgress("error", 0, "不支持的系统，无法自动安装", "")
			return
		}
		switch pkgManager {
		case "apt":
			installCmd = []string{"apt", "install", "-y", "iptables", "iptables-persistent"}
		case "yum":
			installCmd = []string{"yum", "install", "-y", "iptables", "iptables-services"}
		case "dnf":
			installCmd = []string{"dnf", "install", "-y", "iptables", "iptables-services"}
		default:
			sendProgress("error", 0, "无法确定包管理器", "")
			return
		}
		checkCmd = "iptables"
		serviceName = "iptables"

	case "firewalld":
		sendProgress("progress", 5, "准备安装 firewalld 防火墙...", "")
		sendLog("[INFO] firewalld 是 CentOS/RHEL 的默认防火墙工具")
		pkgManager = detectPackageManager(sendLog)
		if pkgManager == "" {
			sendProgress("error", 0, "不支持的系统，无法自动安装", "")
			return
		}
		switch pkgManager {
		case "apt":
			installCmd = []string{"apt", "install", "-y", "firewalld"}
		case "yum":
			installCmd = []string{"yum", "install", "-y", "firewalld"}
		case "dnf":
			installCmd = []string{"dnf", "install", "-y", "firewalld"}
		default:
			sendProgress("error", 0, "无法确定包管理器", "")
			return
		}
		checkCmd = "firewall-cmd"
		serviceName = "firewalld"

	default:
		sendProgress("error", 0, "不支持的防火墙类型: "+firewallType, "")
		return
	}

	// 更新软件包列表 (apt 需要)
	if pkgManager == "apt" {
		sendProgress("progress", 10, "更新软件包列表...", "")
		sendLog("[CMD] apt update")
		_, err := runCommandWithLog("apt", []string{"update"}, sendLog)
		if err != nil {
			sendLog("[WARN] 软件包列表更新失败，继续尝试安装: " + err.Error())
		}
	}

	// 执行安装
	sendProgress("progress", 20, fmt.Sprintf("正在安装 %s...", firewallType), "")
	sendLog(fmt.Sprintf("[CMD] %s", strings.Join(installCmd, " ")))
	_, err := runCommandWithLog(installCmd[0], installCmd[1:], sendLog)
	if err != nil {
		sendLog("[ERROR] 安装失败: " + err.Error())
		sendProgress("error", 0, "安装失败: "+err.Error(), "")
		return
	}
	sendLog("[INFO] 安装命令执行完成")

	// 验证安装
	sendProgress("progress", 70, "验证安装...", "")
	sendLog(fmt.Sprintf("[CMD] which %s", checkCmd))
	_, err = runCommandWithLog("which", []string{checkCmd}, sendLog)
	if err != nil {
		sendLog("[ERROR] 验证失败: 未找到 " + checkCmd)
		sendProgress("error", 0, "安装验证失败，请检查系统环境", "")
		return
	}
	sendLog("[INFO] 验证成功: " + checkCmd + " 已安装")

	// 启用服务
	sendProgress("progress", 80, "配置防火墙服务...", "")
	
	// 根据不同防火墙类型进行初始化配置
	switch firewallType {
	case "ufw":
		// UFW 默认禁用，需要手动启用
		sendLog("[CMD] ufw default deny incoming")
		runCommandWithLog("ufw", []string{"default", "deny", "incoming"}, sendLog)
		sendLog("[CMD] ufw default allow outgoing")
		runCommandWithLog("ufw", []string{"default", "allow", "outgoing"}, sendLog)
		// 允许 SSH 连接，防止被锁在外面
		sendLog("[CMD] ufw allow 22/tcp")
		runCommandWithLog("ufw", []string{"allow", "22/tcp"}, sendLog)
		sendLog("[INFO] UFW 已配置默认规则，SSH 端口已开放")
		
	case "iptables":
		// 启用 iptables 服务
		sendLog("[CMD] systemctl enable iptables")
		runCommandWithLog("systemctl", []string{"enable", "iptables"}, sendLog)
		sendLog("[CMD] systemctl start iptables")
		runCommandWithLog("systemctl", []string{"start", "iptables"}, sendLog)
		
	case "firewalld":
		// 启用 firewalld 服务
		sendLog("[CMD] systemctl enable firewalld")
		runCommandWithLog("systemctl", []string{"enable", "firewalld"}, sendLog)
		sendLog("[CMD] systemctl start firewalld")
		runCommandWithLog("systemctl", []string{"start", "firewalld"}, sendLog)
		// 开放 SSH
		sendLog("[CMD] firewall-cmd --permanent --add-service=ssh")
		runCommandWithLog("firewall-cmd", []string{"--permanent", "--add-service=ssh"}, sendLog)
		sendLog("[CMD] firewall-cmd --reload")
		runCommandWithLog("firewall-cmd", []string{"--reload"}, sendLog)
	}

	sendProgress("progress", 95, "检查服务状态...", "")
	sendLog(fmt.Sprintf("[CMD] systemctl is-active %s", serviceName))
	statusOutput, _ := runCommandWithLog("systemctl", []string{"is-active", serviceName}, sendLog)
	sendLog("[INFO] 服务状态: " + statusOutput)

	sendProgress("complete", 100, fmt.Sprintf("%s 安装完成！", firewallType), "")
}

// detectPackageManager 检测系统包管理器
func detectPackageManager(sendLog func(string)) string {
	// 检测 apt (Debian/Ubuntu)
	if _, err := runCommandWithLog("which", []string{"apt"}, sendLog); err == nil {
		sendLog("[INFO] 检测到包管理器: apt (Debian/Ubuntu)")
		return "apt"
	}
	// 检测 dnf (Fedora/RHEL 8+)
	if _, err := runCommandWithLog("which", []string{"dnf"}, sendLog); err == nil {
		sendLog("[INFO] 检测到包管理器: dnf (Fedora/RHEL 8+)")
		return "dnf"
	}
	// 检测 yum (CentOS/RHEL 7)
	if _, err := runCommandWithLog("which", []string{"yum"}, sendLog); err == nil {
		sendLog("[INFO] 检测到包管理器: yum (CentOS/RHEL)")
		return "yum"
	}
	return ""
}

// runCommandWithLog 执行命令并发送日志
func runCommandWithLog(name string, args []string, sendLog func(string)) (string, error) {
	output, err := firewall.RunCommandWithOutput(name, args...)
	if sendLog != nil && output != "" {
		// 将多行输出拆分为单独的日志行
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				sendLog(line)
			}
		}
	}
	return output, err
}

// UninstallFirewall 卸载防火墙
func (s *FirewallService) UninstallFirewall(firewallType string, keepRules bool, keepPolicies bool, progressChan chan<- dto.UninstallProgress) {
	defer close(progressChan)

	sendProgress := func(progressType string, progress int, message, logContent string) {
		progressChan <- dto.UninstallProgress{
			Type:     progressType,
			Progress: progress,
			Message:  message,
			Log:      logContent,
		}
	}

	sendLog := func(logContent string) {
		sendProgress("log", 0, "", logContent)
	}

	// 检测系统包管理器
	pkgManager := detectPackageManager(sendLog)
	if pkgManager == "" {
		sendProgress("error", 0, "不支持的系统，无法自动卸载", "")
		return
	}

	var serviceName string
	var packages []string
	var rulesBackupDir string

	switch firewallType {
	case "ufw":
		sendProgress("progress", 5, "准备卸载 UFW 防火墙...", "")
		sendLog("[INFO] 即将卸载 UFW 防火墙")
		serviceName = "ufw"
		packages = []string{"ufw"}

	case "iptables":
		sendProgress("progress", 5, "准备卸载 iptables 防火墙...", "")
		sendLog("[INFO] 即将卸载 iptables 防火墙")
		serviceName = "iptables"
		if pkgManager == "apt" {
			packages = []string{"iptables", "iptables-persistent"}
		} else {
			packages = []string{"iptables", "iptables-services"}
		}

	case "firewalld":
		sendProgress("progress", 5, "准备卸载 firewalld 防火墙...", "")
		sendLog("[INFO] 即将卸载 firewalld 防火墙")
		serviceName = "firewalld"
		packages = []string{"firewalld"}

	default:
		sendProgress("error", 0, "不支持的防火墙类型: "+firewallType, "")
		return
	}

	// 备份规则（如果需要保留）
	if keepRules {
		sendProgress("progress", 10, "备份防火墙规则...", "")
		rulesBackupDir = "/etc/gpanel/firewall-backup"
		sendLog(fmt.Sprintf("[CMD] mkdir -p %s", rulesBackupDir))
		runCommandWithLog("mkdir", []string{"-p", rulesBackupDir}, sendLog)

		switch firewallType {
		case "ufw":
			// 备份 UFW 规则
			sendLog("[CMD] cp -r /etc/ufw " + rulesBackupDir + "/ufw")
			runCommandWithLog("cp", []string{"-r", "/etc/ufw", rulesBackupDir + "/ufw"}, sendLog)
			sendLog("[CMD] cp /lib/ufw/user.rules " + rulesBackupDir + "/user.rules")
			runCommandWithLog("cp", []string{"/lib/ufw/user.rules", rulesBackupDir + "/user.rules"}, sendLog)
			sendLog("[CMD] cp /lib/ufw/user6.rules " + rulesBackupDir + "/user6.rules")
			runCommandWithLog("cp", []string{"/lib/ufw/user6.rules", rulesBackupDir + "/user6.rules"}, sendLog)
			sendLog("[INFO] UFW 规则已备份到 " + rulesBackupDir)

		case "iptables":
			// 备份 iptables 规则
			sendLog("[CMD] iptables-save > " + rulesBackupDir + "/iptables.rules")
			runCommandWithLog("sh", []string{"-c", "iptables-save > " + rulesBackupDir + "/iptables.rules"}, sendLog)
			sendLog("[CMD] ip6tables-save > " + rulesBackupDir + "/ip6tables.rules")
			runCommandWithLog("sh", []string{"-c", "ip6tables-save > " + rulesBackupDir + "/ip6tables.rules"}, sendLog)
			sendLog("[INFO] iptables 规则已备份到 " + rulesBackupDir)

		case "firewalld":
			// 备份 firewalld 配置
			sendLog("[CMD] cp -r /etc/firewalld " + rulesBackupDir + "/firewalld")
			runCommandWithLog("cp", []string{"-r", "/etc/firewalld", rulesBackupDir + "/firewalld"}, sendLog)
			sendLog("[INFO] firewalld 配置已备份到 " + rulesBackupDir)
		}
	} else {
		sendProgress("progress", 10, "跳过备份规则...", "")
		sendLog("[INFO] 用户选择不保留规则数据，将清除所有规则")
	}

	// 停止服务
	sendProgress("progress", 20, "停止防火墙服务...", "")
	sendLog(fmt.Sprintf("[CMD] systemctl stop %s", serviceName))
	_, err := runCommandWithLog("systemctl", []string{"stop", serviceName}, sendLog)
	if err != nil {
		sendLog("[WARN] 停止服务失败: " + err.Error())
	}

	// 禁用服务自启动
	sendProgress("progress", 30, "禁用服务自启动...", "")
	sendLog(fmt.Sprintf("[CMD] systemctl disable %s", serviceName))
	_, err = runCommandWithLog("systemctl", []string{"disable", serviceName}, sendLog)
	if err != nil {
		sendLog("[WARN] 禁用服务失败: " + err.Error())
	}

	// 清除规则（如果不保留）
	if !keepRules {
		sendProgress("progress", 40, "清除防火墙规则...", "")
		switch firewallType {
		case "ufw":
			// UFW 重置
			sendLog("[CMD] ufw --force reset")
			runCommandWithLog("ufw", []string{"--force", "reset"}, sendLog)
			sendLog("[INFO] UFW 规则已清除")

		case "iptables":
			// 清除 iptables 规则
			sendLog("[CMD] iptables -F")
			runCommandWithLog("iptables", []string{"-F"}, sendLog)
			sendLog("[CMD] iptables -X")
			runCommandWithLog("iptables", []string{"-X"}, sendLog)
			sendLog("[CMD] iptables -t nat -F")
			runCommandWithLog("iptables", []string{"-t", "nat", "-F"}, sendLog)
			sendLog("[CMD] iptables -t mangle -F")
			runCommandWithLog("iptables", []string{"-t", "mangle", "-F"}, sendLog)
			sendLog("[CMD] iptables -P INPUT ACCEPT")
			runCommandWithLog("iptables", []string{"-P", "INPUT", "ACCEPT"}, sendLog)
			sendLog("[CMD] iptables -P FORWARD ACCEPT")
			runCommandWithLog("iptables", []string{"-P", "FORWARD", "ACCEPT"}, sendLog)
			sendLog("[CMD] iptables -P OUTPUT ACCEPT")
			runCommandWithLog("iptables", []string{"-P", "OUTPUT", "ACCEPT"}, sendLog)
			sendLog("[INFO] iptables 规则已清除")

		case "firewalld":
			// firewalld 的规则会在卸载时自动删除
			sendLog("[INFO] firewalld 规则将在卸载时自动清除")
		}
	}

	// 卸载软件包
	sendProgress("progress", 60, "卸载防火墙软件包...", "")
	var removeCmd []string
	var purgeFlag string

	if pkgManager == "apt" {
		// apt 使用 purge 彻底删除配置
		if !keepPolicies {
			purgeFlag = "--purge"
		}
		removeCmd = append([]string{"remove", "-y", purgeFlag}, packages...)
	} else if pkgManager == "yum" {
		removeCmd = append([]string{"remove", "-y"}, packages...)
	} else if pkgManager == "dnf" {
		removeCmd = append([]string{"remove", "-y"}, packages...)
	}

	sendLog(fmt.Sprintf("[CMD] %s %s", pkgManager, strings.Join(removeCmd, " ")))
	_, err = runCommandWithLog(pkgManager, removeCmd, sendLog)
	if err != nil {
		sendLog("[ERROR] 卸载失败: " + err.Error())
		sendProgress("error", 0, "卸载失败: "+err.Error(), "")
		return
	}
	sendLog("[INFO] 软件包卸载完成")

	// 清理残留配置文件（如果不保留策略）
	if !keepPolicies {
		sendProgress("progress", 80, "清理残留配置...", "")
		switch firewallType {
		case "ufw":
			sendLog("[CMD] rm -rf /etc/ufw")
			runCommandWithLog("rm", []string{"-rf", "/etc/ufw"}, sendLog)
			sendLog("[CMD] rm -rf /lib/ufw")
			runCommandWithLog("rm", []string{"-rf", "/lib/ufw"}, sendLog)
			sendLog("[INFO] UFW 配置文件已清理")

		case "iptables":
			sendLog("[CMD] rm -rf /etc/iptables")
			runCommandWithLog("rm", []string{"-rf", "/etc/iptables"}, sendLog)
			sendLog("[INFO] iptables 配置文件已清理")

		case "firewalld":
			sendLog("[CMD] rm -rf /etc/firewalld")
			runCommandWithLog("rm", []string{"-rf", "/etc/firewalld"}, sendLog)
			sendLog("[INFO] firewalld 配置文件已清理")
		}
	}

	// 验证卸载
	sendProgress("progress", 90, "验证卸载结果...", "")
	var checkCmd string
	switch firewallType {
	case "ufw":
		checkCmd = "ufw"
	case "iptables":
		checkCmd = "iptables"
	case "firewalld":
		checkCmd = "firewall-cmd"
	}

	sendLog(fmt.Sprintf("[CMD] which %s", checkCmd))
	output, err := runCommandWithLog("which", []string{checkCmd}, sendLog)
	if err == nil && output != "" {
		sendLog("[WARN] 检测到 " + checkCmd + " 命令仍然存在，可能卸载不完整")
	} else {
		sendLog("[INFO] 验证成功: " + checkCmd + " 已卸载")
	}

	// 提示备份位置
	if keepRules {
		sendLog("[INFO] 防火墙规则已备份到: " + rulesBackupDir)
		sendLog("[INFO] 重新安装防火墙后可手动恢复规则")
	}

	sendProgress("complete", 100, fmt.Sprintf("%s 卸载完成！", firewallType), "")
}