export interface SSHInfo {
  isActive: boolean;
  port: number;
  listenAddress: string;
  passwordAuth: string;
  pubkeyAuth: string;
  permitRootLogin: string;
  autoStart: boolean;
  useDNS: string;
}

export interface SSHSession {
  pid: number;
  username: string;
  terminal: string;
  host: string;
  loginTime: string;
}

export interface SSHLogItem {
  address: string;
  port: string;
  user: string;
  authMode: string;
  status: string;
  date: string;
}

export interface SSHLogRes {
  total: number;
  items: SSHLogItem[];
}

export interface SSHLogReq {
  page: number;
  pageSize: number;
  status: string;
  info: string;
}

export interface ApiResponse<T = null> {
  code: number;
  message: string;
  data: T;
}
