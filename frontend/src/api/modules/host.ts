import http from '@/utils/axios'
import type {
  HostGroupOperate,
  HostGroupInfo,
  HostGroupSearch,
  HostOperate,
  HostInfo,
  HostSearch,
  HostConnTest,
  HostTreeNode,
  HostMove,
  HostConnInfo,
} from '@/api/interface/host'
import type { ApiResponse } from '@/api/interface/firewall'

// 主机分组 API
export const createGroup = (data: HostGroupOperate) => {
  return http.post<HostGroupInfo>('/api/v1/host-groups', data)
}

export const updateGroup = (id: number, data: HostGroupOperate) => {
  return http.put(`/api/v1/host-groups/${id}`, data)
}

export const deleteGroup = (id: number) => {
  return http.delete(`/api/v1/host-groups/${id}`)
}

export const getGroupByID = (id: number) => {
  return http.get<HostGroupInfo>(`/api/v1/host-groups/${id}`)
}

export const listGroups = (data: HostGroupSearch) => {
  return http.get<{ data: HostGroupInfo[]; total: number }>('/api/v1/host-groups', {
    params: data,
  })
}

// 主机 API
export const createHost = (data: HostOperate) => {
  return http.post<HostInfo>('/api/v1/hosts', data)
}

export const updateHost = (id: number, data: HostOperate) => {
  return http.put(`/api/v1/hosts/${id}`, data)
}

export const deleteHost = (id: number) => {
  return http.delete(`/api/v1/hosts/${id}`)
}

export const getHostByID = (id: number) => {
  return http.get<HostInfo>(`/api/v1/hosts/${id}`)
}

export const listHosts = (data: HostSearch) => {
  return http.get<{ data: HostInfo[]; total: number }>('/api/v1/hosts', {
    params: data,
  })
}

export const testHostConnection = (data: HostConnTest) => {
  return http.post<ApiResponse<{ success: boolean }>>('/api/v1/hosts/test', data)
}

export const moveHosts = (data: HostMove) => {
  return http.post('/api/v1/hosts/move', data)
}

export const getHostTree = () => {
  return http.get<HostTreeNode[]>('/api/v1/hosts/tree')
}

export const getHostForTerminal = (id: number) => {
  return http.get<HostConnInfo>(`/api/v1/hosts/${id}/connection`)
}

export const exportHosts = (encrypted: boolean = true) => {
  return http.get<HostOperate[]>('/api/v1/hosts/export', {
    params: { encrypted },
  })
}

export const importHosts = (data: HostOperate[]) => {
  return http.post<{ success: number; fail: number; message: string }>('/api/v1/hosts/import', data)
}
