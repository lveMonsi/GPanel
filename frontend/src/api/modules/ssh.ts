import request from '@/utils/axios';
import type { ApiResponse, SSHInfo, SSHSession, SSHLogRes, SSHLogReq } from '@/api/interface/ssh';

export const getSSHInfo = () =>
  request.get<ApiResponse<SSHInfo>>('/api/v1/agent/ssh/info').then(res => res.data);

export const operateSSH = (operation: string) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/operate', { operation }).then(res => res.data);

export const updateSSHConfig = (key: string, value: string) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/config', { key, value }).then(res => res.data);

export const getSSHSessions = () =>
  request.get<ApiResponse<SSHSession[]>>('/api/v1/agent/ssh/sessions').then(res => res.data);

export const killSSHSession = (pid: number) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/sessions/kill', { pid }).then(res => res.data);

export const getSSHLogs = (params: SSHLogReq) =>
  request.post<ApiResponse<SSHLogRes>>('/api/v1/agent/ssh/logs', params).then(res => res.data);
