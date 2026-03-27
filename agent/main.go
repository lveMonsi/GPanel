package main

import (
	"gpanel/agent/global"
	"gpanel/agent/middleware"
	"gpanel/agent/models"
	"gpanel/agent/routes"
	"gpanel/agent/service"
	"log"
	"runtime"

	"github.com/gin-gonic/gin"
)

func main() {
	// 显示启动信息
	log.Printf("GPanel Agent starting...")
	log.Printf("Go version: %s, OS/Arch: %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// 初始化配置
	if err := global.InitConfig(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 初始化全局变量
	global.InitGlobals()

	// 初始化数据库
	if err := global.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer global.CloseDB()

	if err := global.DB.AutoMigrate(&models.Setting{}, &models.MonitorBase{}, &models.MonitorIO{}, &models.MonitorNetwork{}, &models.MonitorSetting{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化日志
	global.InitLogger()

	// 启动监控调度器
	service.StartMonitorScheduler()
	defer service.StopMonitorScheduler()

	// 获取服务器配置
	serverMode := global.GetServerMode()
	gin.SetMode(serverMode)

	r := gin.Default()

	// 添加安全中间件
	r.Use(middleware.Security())

	// 注册所有路由
	routes.SetupRouter(r)

	// 获取监听地址
	addr := global.GetListenAddr()
	log.Printf("Starting GPanel Agent on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
}
