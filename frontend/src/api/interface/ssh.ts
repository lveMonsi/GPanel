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
  source: string;
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

export interface SSHFileReq {
  name: 'sshdConf' | 'authKeys';
}

export interface SSHFileUpdateReq {
  key: 'sshdConf' | 'authKeys';
  value: string;
}

export interface SSHKeyInfo {
  id: number;
  createdAt: string;
  name: string;
  mode: 'generate' | 'input' | 'import' | 'sync';
  encryptionMode: 'ed25519' | 'ecdsa' | 'rsa' | 'dsa';
  passPhrase: string;
  description: string;
  publicKey: string;
  privateKey: string;
}

export interface SSHKeyOperate {
  id?: number;
  name: string;
  mode: 'generate' | 'input' | 'import' | 'sync';
  encryptionMode: 'ed25519' | 'ecdsa' | 'rsa' | 'dsa';
  passPhrase: string;
  description: string;
  publicKey: string;
  privateKey: string;
}

export interface SSHKeySearchReq {
  page: number;
  pageSize: number;
}

export interface SSHKeySearchRes {
  total: number;
  items: SSHKeyInfo[];
}

export interface SSHKeyDeleteReq {
  ids: number[];
}

export interface ApiResponse<T = null> {
  code: number;
  message: string;
  data: T;
}
