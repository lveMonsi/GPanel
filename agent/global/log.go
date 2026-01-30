package global

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// InitLogger 初始化日志
func InitLogger() {
	logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: Failed to open log file, using stdout: %v", err)
		logFile = os.Stdout
	}

	// 初始化不同级别的日志记录器
	InfoLogger = log.New(logFile, "[INFO] ", log.LstdFlags|log.Lshortfile)
	ErrorLogger = log.New(logFile, "[ERROR] ", log.LstdFlags|log.Lshortfile)

	if LogLevel == "debug" {
		DebugLogger = log.New(logFile, "[DEBUG] ", log.LstdFlags|log.Lshortfile)
	} else {
		DebugLogger = log.New(os.Stdout, "", 0) // 禁用debug日志
	}
}

// Info 输出信息日志
func Info(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// Error 输出错误日志
func Error(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

// Debug 输出调试日志
func Debug(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}