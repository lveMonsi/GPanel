import request from '@/utils/axios';
import type { ApiResponse } from '@/api/interface/firewall';
import type { OSInfo } from '@/api/interface/system';

// 获取操作系统信息
export const getOSInfo = () => {
  return request.get<ApiResponse<OSInfo>>('/api/v1/system/os').then(res => res.data);
};

// 系统 API 对象
export const systemApi = {
  getOSInfo,
};

export default systemApi;
