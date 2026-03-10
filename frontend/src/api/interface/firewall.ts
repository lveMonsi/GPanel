// 防火墙相关类型定义

// 防火墙基础信息
export interface FirewallBaseInfo {
  name: string;        // 防火墙名称
  isExist: boolean;    // 是否存在
  isActive: boolean;   // 是否激活
  isInit: boolean;     // 是否初始化
  isBind: boolean;     // 是否绑定
  version: string;     // 版本
  pingStatus: string;  // Ping 状态
}

// 防火墙规则信息
export interface FireInfo {
  id?: number;
  protocol: string;    // 协议
  port: string;        // 端口
  address: string;     // IP地址
  strategy: string;    // 策略
  usedStatus?: string; // 使用状态
  description?: string;// 描述
  sourcePort?: string; // 源端口
  targetPort?: string; // 目标端口
  targetIP?: string;   // 目标IP
  chain?: string;      // 链名称
  family?: string;     // IP家族
  num?: string;        // 规则编号
}

// 规则搜索
export interface RuleSearch {
  page: number;
  pageSize: number;
  info?: string;       // 搜索信息
  status?: string;     // 状态筛选
  strategy?: string;   // 策略筛选
  type: string;        // 类型
}

// 防火墙操作
export interface FirewallOperation {
  operation: string;           // 操作类型
  withDockerRestart?: boolean; // 是否重启Docker
}

// 端口规则操作
export interface PortRuleOperate {
  operation: string;   // 操作类型
  port: string;        // 端口
  protocol: string;    // 协议
  strategy: string;    // 策略
  address?: string;    // IP地址
  description?: string;// 描述
}

// 端口规则更新
export interface PortRuleUpdate {
  oldRule: PortRuleOperate;
  newRule: PortRuleOperate;
}

// IP规则操作
export interface IPRuleOperate {
  operation: string;   // 操作类型
  address: string;     // IP地址
  strategy: string;    // 策略
  protocol?: string;   // 协议
  description?: string;// 描述
}

// IP规则更新
export interface IPRuleUpdate {
  oldRule: IPRuleOperate;
  newRule: IPRuleOperate;
}

// 端口转发规则
export interface ForwardRule {
  operation: string;   // 操作类型
  num?: string;        // 规则编号
  protocol: string;    // 协议
  interface?: string;  // 网络接口
  port: string;        // 源端口
  targetIP: string;    // 目标IP
  targetPort: string;  // 目标端口
}

// 端口转发规则操作
export interface ForwardRuleOperate {
  forceDelete: boolean;
  rules: ForwardRule[];
}

// API响应
export interface ApiResponse<T = any> {
  code?: number;
  message?: string;
  data?: T;
}

// 防火墙安装进度消息
export interface InstallProgress {
  type: 'progress' | 'log' | 'error' | 'complete';  // 消息类型
  progress?: number;   // 进度百分比 (0-100)
  message?: string;    // 进度消息
  log?: string;        // 日志内容
}

// 防火墙安装请求
export interface InstallRequest {
  type: 'ufw' | 'iptables' | 'firewalld';  // 防火墙类型
}