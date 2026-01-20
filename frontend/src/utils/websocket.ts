import { ElMessage } from 'element-plus';

export interface WebSocketMessage {
  type: string;
  data: any;
  from?: string;
  to?: string;
}

export type WebSocketEventHandler = (message: WebSocketMessage) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private clientId: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 3000;
  private handlers: Map<string, WebSocketEventHandler[]> = new Map();
  private reconnectTimer: number | null = null;

  constructor(url: string, clientId?: string) {
    this.url = url;
    this.clientId = clientId || this.generateClientId();
  }

  private generateClientId(): string {
    return `client_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // 构建 WebSocket URL
        const wsUrl = `${this.url}?client_id=${this.clientId}`;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log('WebSocket connected:', this.clientId);
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket closed:', event.code, event.reason);
          this.handleReconnect();
        };
      } catch (error) {
        console.error('Failed to create WebSocket:', error);
        reject(error);
      }
    });
  }

  private handleMessage(data: string) {
    try {
      const message: WebSocketMessage = JSON.parse(data);
      console.log('WebSocket message received:', message);

      // 调用所有注册的处理器
      const handlers = this.handlers.get(message.type) || [];
      handlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error('Error in WebSocket handler:', error);
        }
      });

      // 调用通用处理器
      const allHandlers = this.handlers.get('*') || [];
      allHandlers.forEach(handler => {
        try {
          handler(message);
        } catch (error) {
          console.error('Error in WebSocket handler:', error);
        }
      });
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);

      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
      }

      this.reconnectTimer = window.setTimeout(() => {
        this.connect().catch((error) => {
          console.error('Reconnect failed:', error);
        });
      }, this.reconnectDelay);
    } else {
      ElMessage.error('WebSocket 连接失败，请刷新页面重试');
    }
  }

  on(event: string, handler: WebSocketEventHandler) {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, []);
    }
    this.handlers.get(event)!.push(handler);
  }

  off(event: string, handler: WebSocketEventHandler) {
    const handlers = this.handlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  send(message: WebSocketMessage) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  getClientId(): string {
    return this.clientId;
  }
}

// 创建全局 WebSocket 客户端实例
let wsClient: WebSocketClient | null = null;

export function initWebSocket(): WebSocketClient {
  if (!wsClient) {
    // 从当前协议和主机构建 WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const url = `${protocol}//${host}/api/v1/ws`;

    wsClient = new WebSocketClient(url);

    wsClient.connect().catch((error) => {
      console.error('Failed to connect WebSocket:', error);
      ElMessage.error('WebSocket 连接失败');
    });
  }

  return wsClient;
}

export function getWebSocket(): WebSocketClient | null {
  return wsClient;
}

export function destroyWebSocket() {
  if (wsClient) {
    wsClient.disconnect();
    wsClient = null;
  }
}