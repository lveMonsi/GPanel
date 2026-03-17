import request from '@/utils/axios';
import type {
  ApiResponse,
  FileListReq,
  FileListRes,
  FileCreateReq,
  FileDeleteReq,
  FileRenameReq,
  FileMoveReq,
  FileContentReq,
  FileContentRes,
  FileEditReq,
  DirSizeReq,
  DirSizeRes,
  FileCompressReq,
  FileDecompressReq,
  FileChmodReq,
  FileChownReq,
  ProgressInfo,
  FilePreviewReq,
  FilePreviewRes,
  DrivesRes,
  RemoteDownloadReq,
  RemoteDownloadRes,
} from '@/api/interface/file';

export const fileApi = {
  // 获取盘符列表
  getDrives: () => {
    return request.get<ApiResponse<string[]>>('/api/v1/files/drives');
  },

  // 获取文件列表
  getFileList: (data: FileListReq) => {
    return request.post<ApiResponse<FileListRes>>('/api/v1/files/list', data);
  },

  // 创建文件或目录
  createFile: (data: FileCreateReq) => {
    return request.post<ApiResponse>('/api/v1/files/create', data);
  },

  // 删除文件或目录
  deleteFile: (data: FileDeleteReq) => {
    return request.post<ApiResponse>('/api/v1/files/delete', data);
  },

  // 重命名文件或目录
  renameFile: (data: FileRenameReq) => {
    return request.post<ApiResponse>('/api/v1/files/rename', data);
  },

  // 移动或复制文件
  moveFile: (data: FileMoveReq) => {
    return request.post<ApiResponse>('/api/v1/files/move', data);
  },

  // 获取文件内容
  getFileContent: (data: FileContentReq) => {
    return request.post<ApiResponse<FileContentRes>>('/api/v1/files/content', data);
  },

  // 保存文件内容
  saveFileContent: (data: FileEditReq) => {
    return request.post<ApiResponse>('/api/v1/files/save', data);
  },

  // 上传文件
  uploadFile: (path: string, file: File, overwrite = false) => {
    const formData = new FormData();
    formData.append('path', path);
    formData.append('file', file);
    formData.append('overwrite', overwrite.toString());
    return request.post<ApiResponse<{ uploaded: number; total: number }>>(
      '/api/v1/files/upload',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
  },

  // 下载文件
  downloadFile: (path: string) => {
    return request.get(`/api/v1/files/download?path=${encodeURIComponent(path)}`, {
      responseType: 'blob',
    });
  },

  // 获取目录大小
  getDirSize: (data: DirSizeReq) => {
    return request.post<ApiResponse<DirSizeRes>>('/api/v1/files/size', data);
  },

  // 压缩文件
  compressFiles: (data: FileCompressReq) => {
    return request.post<ApiResponse>('/api/v1/files/compress', data);
  },

  // 解压文件
  decompressFile: (data: FileDecompressReq) => {
    return request.post<ApiResponse>('/api/v1/files/decompress', data);
  },

  // 修改文件权限
  chmodFile: (data: FileChmodReq) => {
    return request.post<ApiResponse>('/api/v1/files/chmod', data);
  },

  // 修改文件所有者
  chownFile: (data: FileChownReq) => {
    return request.post<ApiResponse>('/api/v1/files/chown', data);
  },

  // 获取操作进度
  getProgress: (key: string) => {
    return request.get<ApiResponse<ProgressInfo>>(`/api/v1/files/progress?key=${key}`);
  },

  // 预览文件
  previewFile: (data: FilePreviewReq) => {
    return request.post<ApiResponse<FilePreviewRes>>('/api/v1/files/preview', data);
  },

  // 远程下载
  remoteDownload: (data: RemoteDownloadReq) => {
    return request.post<ApiResponse<RemoteDownloadRes>>('/api/v1/files/remote-download', data);
  },
};