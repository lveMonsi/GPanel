package global

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 全局配置变量
var (
	ServerMode       = "release"
	ListenAddr       = "0.0.0.0:9998"
	LogLevel         = "info"
	LogFile          = "/var/log/gpanel/agent.log"
	DataDir          = "/var/lib/gpanel"
	SecurityEntrance = "/"
)

// InitConfig 初始化配置
func InitConfig() error {
	// 读取环境变量
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		ServerMode = mode
	}
	if listen := os.Getenv("GAGENT_LISTEN"); listen != "" {
		ListenAddr = listen
	}
	if logLevel := os.Getenv("GAGENT_LOG_LEVEL"); logLevel != "" {
		LogLevel = logLevel
	}
	if logFile := os.Getenv("GAGENT_LOG_FILE"); logFile != "" {
		LogFile = logFile
	}
	if dataDir := os.Getenv("GAGENT_DATA_DIR"); dataDir != "" {
		DataDir = dataDir
	}

	// 读取配置文件（如果存在）
	configPath := filepath.Join(".", "config", "agent.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		parseConfig(string(data))
	}

	log.Printf("Server mode: %s", ServerMode)
	log.Printf("Listen address: %s", ListenAddr)

	return nil
}

// parseConfig 解析配置文件
func parseConfig(data string) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "server:") || strings.Contains(line, "mode:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				ServerMode = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "listen:") {
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				ListenAddr = strings.TrimSpace(parts[1])
			}
		}
	}
}

// GetServerMode 获取服务器模式
func GetServerMode() string {
	return ServerMode
}

// GetListenAddr 获取监听地址
func GetListenAddr() string {
	return ListenAddr
}

// GetLogLevel 获取日志级别
func GetLogLevel() string {
	return LogLevel
}

// GetLogFile 获取日志文件路径
func GetLogFile() string {
	return LogFile
}

// GetDataDir 获取数据目录
func GetDataDir() string {
	return DataDir
}

// InitGlobals 初始化全局变量
func InitGlobals() {
	// 确保数据目录存在
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		log.Printf("Warning: Failed to create data directory %s: %v", DataDir, err)
	}

	// 确保日志目录存在
	logDir := filepath.Dir(LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: Failed to create log directory %s: %v", logDir, err)
	}
}
