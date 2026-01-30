import request from '@/utils/axios';
import type { ApiResponse, FirewallBaseInfo, RuleSearch, FirewallOperation, PortRuleOperate, PortRuleUpdate, IPRuleOperate, IPRuleUpdate, ForwardRuleOperate } from '@/api/interface/firewall';

// 获取防火墙基础信息
export const loadBaseInfo = () => {
  return request.post<ApiResponse<FirewallBaseInfo>>('/api/v1/agent/firewall/base', {}).then(res => res.data);
};

// 搜索防火墙规则
export const searchRules = (params: RuleSearch) => {
  return request.post<ApiResponse<{ total: number; items: any[] }>>('/api/v1/agent/firewall/search', params).then(res => res.data);
};

// 操作防火墙（启动/停止/重启/Ping防护）
export const operateFirewall = (params: FirewallOperation) => {
  return request.post<ApiResponse>('/api/v1/agent/firewall/operate', params).then(res => res.data);
};

// 操作端口规则
export const operatePortRule = (params: PortRuleOperate) => {
  return request.post<ApiResponse>('/api/v1/agent/firewall/port', params).then(res => res.data);
};

// 更新端口规则
export const updatePortRule = (params: PortRuleUpdate) => {
  return request.post<ApiResponse>('/api/v1/agent/firewall/update/port', params).then(res => res.data);
};

// 操作IP规则
export const operateIPRule = (params: IPRuleOperate) => {
  return request.post<ApiResponse>('/api/v1/agent/firewall/ip', params).then(res => res.data);
};

// 更新IP规则
export const updateIPRule = (params: IPRuleUpdate) => {
  return request.post<ApiResponse>('/api/v1/agent/firewall/update/ip', params).then(res => res.data);
};

// 操作端口转发规则
export const operateForwardRule = (params: ForwardRuleOperate) => {
  return request.post<ApiResponse>('/api/v1/agent/firewall/forward', params).then(res => res.data);
};

// 防火墙 API 对象（可选）
export const firewallApi = {
  loadBaseInfo,
  searchRules,
  operateFirewall,
  operatePortRule,
  updatePortRule,
  operateIPRule,
  updateIPRule,
  operateForwardRule
};

export default firewallApi;