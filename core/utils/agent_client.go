package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"gpanel/global"
)

const (
	// AgentSocketPath Agent Unix Domain Socket 路径
	AgentSocketPath = "/var/run/gpanel/agent.sock"
	// DefaultTimeout 默认超时时间
	DefaultTimeout = 30 * time.Second
)

// AgentClient Agent 客户端
type AgentClient struct {
	baseURL    string
	httpClient *http.Client
	socketPath string
}

// NewAgentClient 创建 Agent 客户端
func NewAgentClient() (*AgentClient, error) {
	var client *AgentClient

	// 优先使用 Unix Domain Socket（仅在 Linux/macOS 上）
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		socketPath := AgentSocketPath
		// 检查 socket 是否存在
		if _, err := os.Stat(socketPath); err == nil {
			client = &AgentClient{
				baseURL:    "http://unix",
				socketPath: socketPath,
				httpClient: &http.Client{
					Timeout: DefaultTimeout,
					Transport: &http.Transport{
						DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
							return net.Dial("unix", socketPath)
						},
					},
				},
			}
		}
	}

	// 如果 Unix Domain Socket 不可用，使用 HTTP
	if client == nil {
		// 从配置缓存获取 Agent 地址
		agentAddr := "localhost:9998"
		if global.ConfigCacheInstance != nil {
			agentAddr = global.ConfigCacheInstance.GetAgentAddress()
		}
		client = &AgentClient{
			baseURL: "http://" + agentAddr,
			httpClient: &http.Client{
				Timeout: DefaultTimeout,
			},
		}
	}

	return client, nil
}

// Request 发送请求到 Agent
func (c *AgentClient) Request(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get 发送 GET 请求
func (c *AgentClient) Get(path string) ([]byte, error) {
	return c.Request("GET", path, nil)
}

// Post 发送 POST 请求
func (c *AgentClient) Post(path string, body interface{}) ([]byte, error) {
	return c.Request("POST", path, body)
}

// IsConnected 检查 Agent 是否连接
func (c *AgentClient) IsConnected() bool {
	_, err := c.Get("/health")
	return err == nil
}