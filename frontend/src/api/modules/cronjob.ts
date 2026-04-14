import http from '@/utils/axios'
import type {
  CronjobCreate,
  CronjobUpdate,
  CronjobDelete,
  CronjobSearch,
  CronjobPageResult,
  CronjobToggle,
  RecordSearch,
  RecordPageResult
} from '@/api/interface/cronjob'

export const createCronjob = (data: CronjobCreate) => {
  return http.post<any>('/api/v1/cronjobs', data)
}

export const updateCronjob = (data: CronjobUpdate) => {
  return http.post<any>('/api/v1/cronjobs/update', data)
}

export const deleteCronjob = (data: CronjobDelete) => {
  return http.post<any>('/api/v1/cronjobs/delete', data)
}

export const searchCronjobs = (data: CronjobSearch) => {
  return http.post<CronjobPageResult>('/api/v1/cronjobs/search', data)
}

export const toggleCronjob = (data: CronjobToggle) => {
  return http.post<any>('/api/v1/cronjobs/toggle', data)
}

export const handleOnceCronjob = (id: number) => {
  return http.post<any>('/api/v1/cronjobs/handle', { id })
}

export const stopCronjob = (id: number) => {
  return http.post<any>('/api/v1/cronjobs/stop', { id })
}

export const searchRecords = (data: RecordSearch) => {
  return http.post<RecordPageResult>('/api/v1/cronjobs/records/search', data)
}

export const getRecordLog = (id: number) => {
  return http.get<any>(`/api/v1/cronjobs/records/${id}/log`)
}

export const cleanRecords = (cronjobId: number) => {
  return http.post<any>('/api/v1/cronjobs/records/clean', { cronjobId })
}

export const getNextExecTimes = (spec: string) => {
  return http.post<any>('/api/v1/cronjobs/next-times', { spec })
}
