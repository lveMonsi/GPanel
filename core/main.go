package main

import (
	"gpanel/global"
	"gpanel/middleware"
	"gpanel/models"
	"gpanel/routes"
	"gpanel/service"
	"log"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	Version    = "dev"
	BuildTime  = "unknown"
	GitCommit  = "unknown"
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

	// 初始化配置热重载（每30秒检查一次）
	global.InitConfigReloader(30 * time.Second)

	// 从配置缓存获取服务器配置
	serverMode := global.ConfigCacheInstance.GetServerMode()
	gin.SetMode(serverMode)

	r := gin.Default()

	// 添加安全入口中间件
	r.Use(middleware.SecurityEntrance())

	// 添加安全中间件
	r.Use(middleware.Security())

	SetupFrontend(r)
	routes.SetupRouter(r)

	// 从配置缓存获取端口配置
	serverPort := global.ConfigCacheInstance.GetServerPort()

	// 从配置缓存获取安全入口配置
	securityEntrance := global.ConfigCacheInstance.GetSecurityEntrance()

	addr := ":" + serverPort
	log.Printf("Starting GPanel server on %s", addr)
	log.Printf("Security entrance: %s", securityEntrance)
	log.Printf("Language: %s, Timezone: %s", global.ConfigCacheInstance.GetLanguage(), global.ConfigCacheInstance.GetTimezone())
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}