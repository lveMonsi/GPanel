import { TerminalConfig, SSHConfig } from '../interface/terminal';

export function getLocalTerminalUrl(config: TerminalConfig): string {
  // 使用 Core 服务（同源），Core 服务会代理到 Agent
  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
  const host = window.location.host;
  return `${protocol}://${host}/api/v1/terminal/local?cols=${config.cols}&rows=${config.rows}`;
}

export function getSSHTerminalUrl(config: TerminalConfig): string {
  // 使用 Core 服务（同源），Core 服务会代理到 Agent
  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
  const host = window.location.host;
  const params = new URLSearchParams({
    cols: config.cols.toString(),
    rows: config.rows.toString()
  });
  return `${protocol}://${host}/api/v1/terminal/ssh?${params.toString()}`;
}

export const terminalApi = {
  // 获取本地终端 URL
  getLocalTerminalUrl: (config: TerminalConfig): string => {
    return getLocalTerminalUrl(config);
  },

  // 获取 SSH 终端 URL（不包含凭证）
  getSSHTerminalUrl: (config: TerminalConfig): string => {
    return getSSHTerminalUrl(config);
  }
};