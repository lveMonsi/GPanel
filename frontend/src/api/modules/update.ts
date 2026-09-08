import request from '@/utils/axios'
import type { ApiEnvelope, BuildInfo, UpdateRequest, UpdateStatus } from '@/api/interface/update'

export const getBuildInfo = () =>
  request.get<ApiEnvelope<BuildInfo>>('/api/v1/version').then(response => response.data.build as BuildInfo)

export const getUpdateStatus = () =>
  request.get<ApiEnvelope<UpdateStatus>>('/api/v1/update/status').then(response => response.data.status as UpdateStatus)

export const stageUpdate = (data: UpdateRequest) =>
  request.post<ApiEnvelope<UpdateStatus>>('/api/v1/update', data).then(response => response.data.status as UpdateStatus)

export const applyUpdate = () =>
  request.post<ApiEnvelope<UpdateStatus>>('/api/v1/update/apply').then(response => response.data.status as UpdateStatus)

export const updateApi = { getBuildInfo, getUpdateStatus, stageUpdate, applyUpdate }
export default updateApi
