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
		output, err := runCommandWithLog("apt", "update", sendLog)
		if err != nil {
			sendLog("[WARN] 软件包列表更新失败，继续尝试安装: " + err.Error())
		} else {
			sendLog(output)
		}
	}

	// 执行安装
	sendProgress("progress", 20, fmt.Sprintf("正在安装 %s...", firewallType), "")
	sendLog(fmt.Sprintf("[CMD] %s", strings.Join(installCmd, " ")))
	output, err := runCommandWithLog(installCmd[0], installCmd[1:]..., sendLog)
	if err != nil {
		sendLog("[ERROR] 安装失败: " + err.Error())
		sendProgress("error", 0, "安装失败: "+err.Error(), "")
		return
	}
	sendLog(output)
	sendLog("[INFO] 安装命令执行完成")

	// 验证安装
	sendProgress("progress", 70, "验证安装...", "")
	sendLog(fmt.Sprintf("[CMD] which %s", checkCmd))
	_, err = runCommandWithLog("which", checkCmd, sendLog)
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
		runCommandWithLog("ufw", "default", "deny", "incoming", sendLog)
		sendLog("[CMD] ufw default allow outgoing")
		runCommandWithLog("ufw", "default", "allow", "outgoing", sendLog)
		// 允许 SSH 连接，防止被锁在外面
		sendLog("[CMD] ufw allow 22/tcp")
		runCommandWithLog("ufw", "allow", "22/tcp", sendLog)
		sendLog("[INFO] UFW 已配置默认规则，SSH 端口已开放")
		
	case "iptables":
		// 启用 iptables 服务
		sendLog("[CMD] systemctl enable iptables")
		runCommandWithLog("systemctl", "enable", "iptables", sendLog)
		sendLog("[CMD] systemctl start iptables")
		runCommandWithLog("systemctl", "start", "iptables", sendLog)
		
	case "firewalld":
		// 启用 firewalld 服务
		sendLog("[CMD] systemctl enable firewalld")
		runCommandWithLog("systemctl", "enable", "firewalld", sendLog)
		sendLog("[CMD] systemctl start firewalld")
		runCommandWithLog("systemctl", "start", "firewalld", sendLog)
		// 开放 SSH
		sendLog("[CMD] firewall-cmd --permanent --add-service=ssh")
		runCommandWithLog("firewall-cmd", "--permanent", "--add-service=ssh", sendLog)
		sendLog("[CMD] firewall-cmd --reload")
		runCommandWithLog("firewall-cmd", "--reload", sendLog)
	}

	sendProgress("progress", 95, "检查服务状态...", "")
	sendLog(fmt.Sprintf("[CMD] systemctl status %s", serviceName))
	statusOutput, err := runCommandWithLog("systemctl", "is-active", serviceName, sendLog)
	if err != nil {
		sendLog("[WARN] 服务状态检查: " + statusOutput)
	} else {
		sendLog("[INFO] 服务状态: " + statusOutput)
	}

	sendProgress("complete", 100, fmt.Sprintf("%s 安装完成！", firewallType), "")
}

// detectPackageManager 检测系统包管理器
func detectPackageManager(sendLog func(string)) string {
	// 检测 apt (Debian/Ubuntu)
	if _, err := runCommandWithLog("which", "apt", sendLog); err == nil {
		sendLog("[INFO] 检测到包管理器: apt (Debian/Ubuntu)")
		return "apt"
	}
	// 检测 dnf (Fedora/RHEL 8+)
	if _, err := runCommandWithLog("which", "dnf", sendLog); err == nil {
		sendLog("[INFO] 检测到包管理器: dnf (Fedora/RHEL 8+)")
		return "dnf"
	}
	// 检测 yum (CentOS/RHEL 7)
	if _, err := runCommandWithLog("which", "yum", sendLog); err == nil {
		sendLog("[INFO] 检测到包管理器: yum (CentOS/RHEL)")
		return "yum"
	}
	return ""
}

// runCommandWithLog 执行命令并发送日志
func runCommandWithLog(name string, args ...string) (string, error) {
	return firewall.RunCommandWithOutput(name, args...)
}