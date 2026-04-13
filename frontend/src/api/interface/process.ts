// 进程搜索请求
export interface ProcessSearchReq {
  pid?: number
  name?: string
  username?: string
}

// 网络连接搜索请求
export interface NetSearchReq {
  processID?: number
  processName?: string
  port?: number
}

// 终止进程请求
export interface ProcessStopReq {
  PID: number
}

// 进程列表项
export interface ProcessInfo {
  PID: number
  name: string
  PPID: number
  username: string
  status: string
  startTime: string
  numThreads: number
  numConnections: number
  cpuPercent: string
  cpuValue: number
  rss: string
  rssValue: number
}

// 进程详情
export interface ProcessDetail {
  PID: number
  name: string
  PPID: number
  username: string
  status: string
  startTime: string
  numThreads: number
  numConnections: number
  cpuPercent: string
  cpuValue: number

  diskRead: string
  diskWrite: string
  cmdLine: string

  rss: string
  rssValue: number
  vms: string
  hwm: string
  data: string
  stack: string
  locked: string
  swap: string
  dirty: string
  pss: string
  uss: string
  shared: string
  text: string

  envs: string[]
  openFiles: OpenFileStat[]
  connects: NetConnection[]
}

// 打开文件信息
export interface OpenFileStat {
  path: string
  fd: number
}

// 网络连接信息
export interface NetConnection {
  type: string
  status: string
  localaddr: string
  remoteaddr: string
  PID: number
  name: string
}
