import http from '@/utils/axios'
import type {
  OperationLogSearch,
  OperationLogPageResult,
  OperationLogClean,
  OperationLogStats,
  SystemLogSearch,
  SystemLogPageResult,
  SystemLogInfo
} from '@/api/interface/log'

export const searchOperationLogs = (data: OperationLogSearch) => {
  return http.post<{ data: OperationLogPageResult }>('/api/v1/logs/operation/search', data)
}

export const cleanOperationLogs = (data: OperationLogClean) => {
  return http.post<{ data: { deleted: number } }>('/api/v1/logs/operation/clean', data)
}

export const getOperationLogStats = () => {
  return http.get<{ data: OperationLogStats }>('/api/v1/logs/operation/stats')
}

export const searchSystemLogs = (data: SystemLogSearch) => {
  return http.post<{ data: SystemLogPageResult }>('/api/v1/logs/system/search', data)
}

export const getSystemLogInfo = () => {
  return http.get<{ data: SystemLogInfo }>('/api/v1/logs/system/info')
}
