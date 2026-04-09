package main

import (
	"gpanel/dto"
	"gpanel/global"
	"gpanel/middleware"
	"gpanel/models"
	"gpanel/routes"
	"gpanel/service"
	"log"
	"runtime"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := global.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer global.CloseLogger()

	// 显示启动信息
	log.Printf("GPanel starting...")
	log.Printf("Go version: %s, OS/Arch: %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// 初始化环境变量配置（JWT 密钥、Agent API Key 等）
	global.InitEnvConfig()

	// 初始化数据库
	if err := global.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer global.CloseDB()

	// 自动迁移数据库表
	if err := global.DB.AutoMigrate(
		&models.Setting{},
		&models.HostGroup{},
		&models.Host{},
		&models.SessionHistory{},
		&models.QuickCommand{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化默认设置
	settingService := service.NewSettingService()
	if err := settingService.InitializeDefaultSettings(); err != nil {
		log.Printf("Warning: Failed to initialize default settings: %v", err)
	}

	// 初始化默认主机分组
	hostService := service.NewHostService()
	if err := initializeDefaultHostGroups(hostService); err != nil {
		log.Printf("Warning: Failed to initialize default host groups: %v", err)
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

	// 创建 Gin 引擎，优化性能
	r := gin.New()

	// 使用自定义的 Logger 和 Recovery 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 设置信任的代理
	r.SetTrustedProxies(nil)

	// 优化性能：禁用不必要的功能
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// 添加安全入口中间件
	r.Use(middleware.SecurityEntrance())

	// 添加安全中间件
	r.Use(middleware.Security())

	// 注册所有路由（包括前端静态文件服务）
	routes.SetupRouter(r)

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

// initializeDefaultHostGroups 初始化默认主机分组
func initializeDefaultHostGroups(hostService service.IHostService) error {
	// 检查是否已存在默认分组
	groups, _, err := hostService.ListGroups(dto.HostGroupSearch{
		PageInfo: dto.PageInfo{Page: 1, PageSize: 10},
	})
	if err != nil {
		return err
	}

	// 如果没有任何分组，创建默认分组
	if len(groups) == 0 {
		_, err := hostService.CreateGroup(dto.HostGroupOperate{
			Name:        "默认分组",
			Description: "系统默认创建的主机分组",
		})
		return err
	}

	return nil
}
