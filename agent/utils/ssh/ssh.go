package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// ConnInfo SSH 连接信息
type ConnInfo struct {
	User        string        `json:"user"`
	Addr        string        `json:"addr"`
	Port        int           `json:"port"`
	AuthMode    string        `json:"authMode"`
	Password    string        `json:"password"`
	PrivateKey  []byte        `json:"privateKey"`
	PassPhrase  []byte        `json:"passPhrase"`
	DialTimeOut time.Duration `json:"dialTimeOut"`
}

// SSHClient SSH 客户端
type SSHClient struct {
	Client *gossh.Client `json:"client"`
}

var (
	knownHostsFile string
	knownHostsInit sync.Once
)

// getKnownHostsFile 获取已知主机文件路径
func getKnownHostsFile() string {
	knownHostsInit.Do(func() {
		// 使用应用数据目录存储已知主机
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/tmp"
		}
		gpanelDir := filepath.Join(homeDir, ".gpanel")
		if err := os.MkdirAll(gpanelDir, 0700); err != nil {
			fmt.Printf("[WARNING] Failed to create gpanel directory %s: %v\n", gpanelDir, err)
		}
		knownHostsFile = filepath.Join(gpanelDir, "known_hosts")
	})
	return knownHostsFile
}

// createHostKeyCallback 创建主机密钥回调函数
// 支持首次连接时自动添加主机密钥（类似 OpenSSH 的 StrictHostKeyChecking=ask）
// 注意：在生产环境中，应该要求用户确认主机密钥指纹
func createHostKeyCallback() gossh.HostKeyCallback {
	knownHostsPath := getKnownHostsFile()

	// 创建 knownhosts 回调
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		// 如果 known_hosts 文件不存在或无法创建，回退到不安全模式（仅用于开发）
		// 生产环境应该要求用户提供 known_hosts 文件或手动验证主机密钥
		// 记录警告日志
		fmt.Printf("[WARNING] Failed to create known_hosts callback: %v. Using insecure host key verification.\n", err)
		return gossh.InsecureIgnoreHostKey()
	}

	// 包装回调以支持首次连接
	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		err := callback(hostname, remote, key)
		if err == nil {
			return nil
		}

		// 检查是否是未知主机错误
		if keyErr, ok := err.(*knownhosts.KeyError); ok {
			// 检查是否是主机密钥变更（严重的安全风险）
			if len(keyErr.Want) > 0 {
				// 主机密钥已变更，拒绝连接（防止 MITM 攻击）
				return fmt.Errorf("host key for %s has changed! This could indicate a man-in-the-middle attack.\nWant: %v\nGot: %v",
					hostname, keyErr.Want, key)
			}

			// 首次连接，自动添加主机密钥
			// 注意：在生产环境中，这应该要求用户确认
			// 记录主机密钥指纹用于审计
			fingerprint := gossh.FingerprintSHA256(key)
			fmt.Printf("[INFO] Adding new host key for %s (SHA256: %s)\n", hostname, fingerprint)

			file, ferr := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
			if ferr != nil {
				return fmt.Errorf("failed to open known_hosts file: %w", ferr)
			}
			defer file.Close()

			line := knownhosts.Line([]string{hostname}, key)
			if _, werr := fmt.Fprintln(file, line); werr != nil {
				return fmt.Errorf("failed to write to known_hosts file: %w", werr)
			}

			return nil
		}

		return err
	}
}

// NewClient 创建新的 SSH 客户端
func NewClient(c ConnInfo) (*SSHClient, error) {
	config := &gossh.ClientConfig{}
	config.SetDefaults()

	addr := net.JoinHostPort(c.Addr, fmt.Sprintf("%d", c.Port))
	config.User = c.User

	if c.AuthMode == "password" {
		config.Auth = []gossh.AuthMethod{gossh.Password(c.Password)}
	} else {
		signer, err := makePrivateKeySigner(c.PrivateKey, c.PassPhrase)
		if err != nil {
			return nil, fmt.Errorf("failed to create signer: %w", err)
		}
		config.Auth = []gossh.AuthMethod{gossh.PublicKeys(signer)}
	}

	if c.DialTimeOut == 0 {
		c.DialTimeOut = 5 * time.Second
	}
	config.Timeout = c.DialTimeOut

	// 使用主机密钥验证
	config.HostKeyCallback = createHostKeyCallback()

	proto := "tcp"
	if strings.Contains(c.Addr, ":") {
		proto = "tcp6"
	}

	fmt.Printf("[INFO] Connecting to SSH server: %s (user: %s, mode: %s)\n", addr, c.User, c.AuthMode)
	client, err := gossh.Dial(proto, addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server %s: %w", addr, err)
	}

	fmt.Printf("[INFO] SSH connection established: %s\n", addr)
	return &SSHClient{Client: client}, nil
}

// Close 关闭 SSH 连接
func (c *SSHClient) Close() {
	_ = c.Client.Close()
}

// makePrivateKeySigner 创建私钥签名器
func makePrivateKeySigner(privateKey []byte, passPhrase []byte) (gossh.Signer, error) {
	if len(passPhrase) != 0 {
		return gossh.ParsePrivateKeyWithPassphrase(privateKey, passPhrase)
	}
	return gossh.ParsePrivateKey(privateKey)
}