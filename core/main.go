package main

import (
	"gpanel/global"
	"gpanel/middleware"
	"gpanel/models"
	"gpanel/routes"
	"gpanel/service"
	"log"
	"runtime"

	"github.com/gin-gonic/gin"
)

var (
	Version   = "v0.0.1"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 显示版本信息
	log.Printf("GPanel v%s (commit: %s, built: %s)", Version, GitCommit, BuildTime)
	log.Printf("Go version: %s, OS/Arch: %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// 初始化数据库
	if err := global.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer global.CloseDB()

	// 自动迁移数据库表
	if err := global.DB.AutoMigrate(
		&models.Setting{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化默认设置
	settingService := service.NewSettingService()
	if err := settingService.InitializeDefaultSettings(); err != nil {
		log.Printf("Warning: Failed to initialize default settings: %v", err)
	}

	// 初始化配置缓存
	if err := global.InitConfigCache(); err != nil {
		log.Fatalf("Failed to initialize config cache: %v", err)
	}

	// 初始化配置热重载器（禁用自动重载，仅支持手动触发）
	global.InitConfigReloader(0)

	// 从配置缓存获取服务器配置
	serverMode := global.ConfigCacheInstance.GetServerMode()
	gin.SetMode(serverMode)

	r := gin.Default()

	// 添加安全入口中间件
	r.Use(middleware.SecurityEntrance())

	// 添加安全中间件
	r.Use(middleware.Security())

	// 先注册 API 路由
	routes.SetupRouter(r)

	// 最后注册前端路由（通配符路由）
	SetupFrontend(r)

	// 从配置缓存获取端口配置
	serverPort := global.ConfigCacheInstance.GetServerPort()

	// 从配置缓存获取安全入口配置
	securityEntrance := global.ConfigCacheInstance.GetSecurityEntrance()

	// 输出安全入口到控制台
	if securityEntrance != "/" {
		log.Printf("========================================")
		log.Printf("  Security Entrance: %s", securityEntrance)
		log.Printf("========================================")
	}

	addr := ":" + serverPort
	log.Printf("Starting GPanel server on %s", addr)
	log.Printf("Security entrance: %s", securityEntrance)
	log.Printf("Language: %s, Timezone: %s", global.ConfigCacheInstance.GetLanguage(), global.ConfigCacheInstance.GetTimezone())
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
