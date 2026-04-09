import request from '@/utils/axios';
import type {
  ApiResponse,
  SSHInfo,
  SSHSession,
  SSHLogRes,
  SSHLogReq,
  SSHFileReq,
  SSHFileUpdateReq,
  SSHKeyOperate,
  SSHKeySearchReq,
  SSHKeySearchRes,
  SSHKeyDeleteReq,
} from '@/api/interface/ssh';

export const getSSHInfo = () =>
  request.get<ApiResponse<SSHInfo>>('/api/v1/agent/ssh/info').then(res => res.data);

export const operateSSH = (operation: string) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/operate', { operation }).then(res => res.data);

export const updateSSHConfig = (key: string, value: string) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/config', { key, value }).then(res => res.data);

export const getSSHFile = (name: SSHFileReq['name']) =>
  request.post<ApiResponse<string>>('/api/v1/agent/ssh/file', { name }).then(res => res.data);

export const updateSSHFile = (key: SSHFileUpdateReq['key'], value: string) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/file/update', { key, value }).then(res => res.data);

export const getSSHKeys = (params: SSHKeySearchReq) =>
  request.post<ApiResponse<SSHKeySearchRes>>('/api/v1/agent/ssh/keys/search', params).then(res => res.data);

export const createSSHKey = (data: SSHKeyOperate) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/keys', data).then(res => res.data);

export const updateSSHKey = (data: SSHKeyOperate) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/keys/update', data).then(res => res.data);

export const deleteSSHKeys = (data: SSHKeyDeleteReq) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/keys/delete', data).then(res => res.data);

export const syncSSHKeys = () =>
  request.post<ApiResponse>('/api/v1/agent/ssh/keys/sync').then(res => res.data);

export const getSSHSessions = () =>
  request.get<ApiResponse<SSHSession[]>>('/api/v1/agent/ssh/sessions').then(res => res.data);

export const killSSHSession = (pid: number) =>
  request.post<ApiResponse>('/api/v1/agent/ssh/sessions/kill', { pid }).then(res => res.data);

export const getSSHLogs = (params: SSHLogReq) =>
  request.post<ApiResponse<SSHLogRes>>('/api/v1/agent/ssh/logs', params).then(res => res.data);
