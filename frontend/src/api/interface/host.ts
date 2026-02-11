// 主机分组操作
export interface HostGroupOperate {
  id?: number
  name: string
  description?: string
}

// 主机分组信息
export interface HostGroupInfo {
  id: number
  createdAt: string
  updatedAt: string
  name: string
  description: string
  hostCount: number
}

// 主机分组搜索
export interface HostGroupSearch {
  page: number
  pageSize: number
  info?: string
}

// 主机操作
export interface HostOperate {
  id?: number
  groupID: number
  name: string
  addr: string
  port: number
  user: string
  authMode: 'password' | 'key'
  password?: string
  privateKey?: string
  passPhrase?: string
  rememberPassword?: boolean
  description?: string
}

// 主机信息
export interface HostInfo {
  id: number
  createdAt: string
  updatedAt: string
  groupID: number
  groupName: string
  name: string
  addr: string
  port: number
  user: string
  authMode: 'password' | 'key'
  rememberPassword: boolean
  description: string
}

// 主机搜索
export interface HostSearch {
  page: number
  pageSize: number
  groupID?: number
  info?: string
}

// 主机连接测试
export interface HostConnTest {
  addr: string
  port: number
  user: string
  authMode: 'password' | 'key'
  password?: string
  privateKey?: string
  passPhrase?: string
}

// 主机树节点
export interface HostTreeNode {
  id: number
  name: string
  type: 'group' | 'host'
  children?: HostTreeNode[]
}

// 主机移动
export interface HostMove {
  hostIDs: number[]
  groupID: number
}

// 主机连接信息（用于终端）
export interface HostConnInfo {
  id: number
  groupID: number
  name: string
  addr: string
  port: number
  user: string
  authMode: 'password' | 'key'
  password: string
  privateKey: string
  passPhrase: string
  rememberPassword: boolean
  description: string
}