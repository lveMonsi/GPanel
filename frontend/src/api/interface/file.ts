export interface FileInfo {
  path: string;
  name: string;
  size: number;
  isDir: boolean;
  isSymlink: boolean;
  isHidden: boolean;
  linkPath: string;
  extension: string;
  mode: string;
  mimeType: string;
  modTime: string;
  user: string;
  group: string;
  uid: string;
  gid: string;
}

export interface FileListReq {
  path: string;
  search?: string;
  showHidden?: boolean;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: string;
}

export interface FileListRes {
  path: string;
  name: string;
  items: FileInfo[];
  itemTotal: number;
}

export interface FileCreateReq {
  path: string;
  isDir: boolean;
  mode?: number;
}

export interface FileDeleteReq {
  path: string;
  force?: boolean;
}

export interface FileRenameReq {
  oldPath: string;
  newPath: string;
}

export interface FileMoveReq {
  oldPaths: string[];
  newPath: string;
  type: 'cut' | 'copy';
  cover?: boolean;
}

export interface FileContentReq {
  path: string;
}

export interface FileContentRes {
  path: string;
  name: string;
  content: string;
  mode: string;
  size: number;
  mimeType: string;
  truncated: boolean;  // 文件是否被截断（超过大小限制）
}

export interface FileEditReq {
  path: string;
  content: string;
}

export interface DirSizeReq {
  path: string;
}

export interface DirSizeRes {
  path: string;
  size: number;
}

export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

export interface ProgressInfo {
  key: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  percent: number;
  message: string;
  total: number;
  processed: number;
}

export interface FilePreviewReq {
  path: string;
}

export interface FilePreviewRes {
  path: string;
  name: string;
  type: 'image' | 'text' | 'video' | 'audio' | 'pdf' | 'none';
  content: string;
  size: number;
}

export interface FileCompressReq {
  files: string[];
  dst: string;
  name: string;
  type?: 'zip' | 'tar.gz';
}

export interface FileDecompressReq {
  path: string;
  dst: string;
  type?: 'zip' | 'tar.gz';
}

export interface FileChmodReq {
  path: string;
  mode: number;
  sub?: boolean;
}

export interface FileChownReq {
  path: string;
  user?: string;
  group?: string;
  sub?: boolean;
}

export interface DrivesRes {
  drives: string[];
}