// 操作日志
export interface OperationLog {
  id: number
  createdAt: string
  username: string
  ip: string
  resource: string
  action: string
  detail: string
  status: string
}

export interface OperationLogSearch {
  page: number
  pageSize: number
  username?: string
  resource?: string
  action?: string
  status?: string
  keyword?: string
  startTime?: string
  endTime?: string
}

export interface OperationLogClean {
  retainDays: number
}

export interface OperationLogStats {
  total: number
  todayCount: number
  resourceStat: Record<string, number>
}

export interface OperationLogPageResult {
  items: OperationLog[]
  total: number
}

// 系统日志
export interface SystemLogSearch {
  page: number
  pageSize: number
  keyword?: string
  logFile?: string
}

export interface SystemLogPageResult {
  lines: string[]
  total: number
}

export interface SystemLogFile {
  name: string
  path: string
  size: number
}

export interface SystemLogInfo {
  files: SystemLogFile[]
}
