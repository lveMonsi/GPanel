import http from '@/utils/axios';
import type {
  QuickCommand,
  QuickCommandCreate,
  QuickCommandUpdate,
  QuickCommandDelete,
  QuickCommandSearch,
  QuickCommandPageResult,
  CommandTreeItem
} from '@/api/interface/quick_command';

// 创建快速命令
export const createQuickCommand = (data: QuickCommandCreate) => {
  return http.post<any>('/api/v1/quick-commands', data);
};

// 更新快速命令
export const updateQuickCommand = (data: QuickCommandUpdate) => {
  return http.post<any>('/api/v1/quick-commands/update', data);
};

// 删除快速命令
export const deleteQuickCommand = (data: QuickCommandDelete) => {
  return http.post<any>('/api/v1/quick-commands/delete', data);
};

// 搜索快速命令
export const searchQuickCommands = (data: QuickCommandSearch) => {
  return http.post<QuickCommandPageResult>('/api/v1/quick-commands/search', data);
};

// 获取所有快速命令
export const getAllQuickCommands = () => {
  return http.get<QuickCommand[]>('/api/v1/quick-commands/all');
};

// 获取命令树
export const getCommandTree = () => {
  return http.get<CommandTreeItem[]>('/api/v1/quick-commands/tree');
};