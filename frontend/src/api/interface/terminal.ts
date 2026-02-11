export interface TerminalConfig {
  cols: number;
  rows: number;
  fontSize: number;
  fontFamily: string;
  theme: 'light' | 'dark';
}

export interface SSHConfig {
  host: string;
  port: number;
  user: string;
  password: string;
  authMode: 'password' | 'key';
  key?: string;
}

export interface WsMsg {
  type: 'cmd' | 'resize' | 'heartbeat' | 'connect';
  data?: string;
  cols?: number;
  rows?: number;
  timestamp?: number;
}

export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}