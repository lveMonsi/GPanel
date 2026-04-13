import request from '@/utils/axios'
import type { ApiResponse } from '@/api/interface/firewall'
import type {
  ProcessSearchReq,
  NetSearchReq,
  ProcessStopReq,
  ProcessInfo,
  ProcessDetail,
  NetConnection,
} from '@/api/interface/process'

// 获取进程列表
export const listProcesses = (params: ProcessSearchReq) => {
  return request.post<ApiResponse<ProcessInfo[]>>('/api/v1/agent/process/list', params)
}

// 获取进程详情
export const getProcessDetail = (pid: number) => {
  return request.get<ApiResponse<ProcessDetail>>(`/api/v1/agent/process/${pid}`)
}

// 终止进程
export const stopProcess = (params: ProcessStopReq) => {
  return request.post<ApiResponse<null>>('/api/v1/agent/process/stop', params)
}

// 获取网络连接列表
export const listNetConnections = (params: NetSearchReq) => {
  return request.post<ApiResponse<NetConnection[]>>('/api/v1/agent/process/net', params)
}
