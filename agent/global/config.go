package global

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 全局配置变量（带默认值）
var (
	ServerMode      = "release"
	ListenAddr      = "0.0.0.0:9998"
	LogLevel        = "info"
	LogFile         = "/var/log/gpanel/agent.log"
	DataDir         = "/var/lib/gpanel"
	AllowedOrigins  = []string{} // 允许的 WebSocket Origins，为空时允许所有
)

// InitConfig 初始化配置
// Agent 配置完全通过环境变量控制，不再使用配置文件
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
	// 读取允许的 Origins（逗号分隔）
	if origins := os.Getenv("GAGENT_ALLOWED_ORIGINS"); origins != "" {
		AllowedOrigins = strings.Split(origins, ",")
		for i, o := range AllowedOrigins {
			AllowedOrigins[i] = strings.TrimSpace(o)
		}
	}

	log.Printf("Server mode: %s", ServerMode)
	log.Printf("Listen address: %s", ListenAddr)
	log.Printf("Data directory: %s", DataDir)
	if len(AllowedOrigins) > 0 {
		log.Printf("Allowed origins: %v", AllowedOrigins)
	}

	return nil
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

// GetAllowedOrigins 获取允许的 WebSocket Origins
func GetAllowedOrigins() []string {
	return AllowedOrigins
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