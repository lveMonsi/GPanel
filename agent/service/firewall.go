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
	if s.client == nil {
		return dto.FirewallBaseInfo{
			Name:    "-",
			IsExist: false,
		}, nil
	}

	var baseInfo dto.FirewallBaseInfo
	baseInfo.Version = "-"
	baseInfo.Name = s.client.Name()

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
		baseInfo.IsExist = true
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