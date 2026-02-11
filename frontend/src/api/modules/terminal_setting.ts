import http from '@/utils/axios';
import type { TerminalInfo, TerminalUpdate } from '@/api/interface/terminal_setting';

// 获取终端设置
export const getTerminalSetting = () => {
  return http.get<TerminalInfo>('/api/v1/settings/terminal');
};

// 更新终端设置
export const updateTerminalSetting = (data: TerminalUpdate) => {
  return http.post<any>('/api/v1/settings/terminal', data);
};