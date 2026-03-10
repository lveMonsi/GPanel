package routes

import (
	"embed"
	"gpanel/controllers"
	"gpanel/middleware"
	"gpanel/utils"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var frontendFS embed.FS

func SetupRouter(r *gin.Engine) {
	// 启动 WebSocket 集线器
	go utils.Hub.Run()

	// 启动进度清理任务
	utils.StartCleanupTask()

	// 测试路由
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "Test route works!")
	})

	// 调试路由 - 检查安全入口中间件
	r.GET("/debug-login", func(c *gin.Context) {
		c.String(200, "Debug login route - should be accessible")
	})

	// 登录路由（不需要 /api 前缀）

		r.GET("/login", func(c *gin.Context) {

			// 使用 embed.FS 返回前端登录页面

			frontendDist, err := fs.Sub(frontendFS, "web/dist")

			if err != nil {

				c.String(500, "Failed to load frontend: %v", err)

				return

			}

			indexFile, err := frontendDist.Open("index.html")

			if err != nil {

				c.String(500, "Failed to open index.html: %v", err)

				return

			}

			defer indexFile.Close()

			

			// 读取文件内容

			content, err := io.ReadAll(indexFile)

			if err != nil {

				c.String(500, "Failed to read index.html: %v", err)

				return

			}

			

			c.Header("Content-Type", "text/html; charset=utf-8")

			c.Data(200, "text/html; charset=utf-8", content)

		})

		r.POST("/login", controllers.Login)

	// 设置前端静态文件服务（使用 frontend.go 中的 embed.FS 版本）
	SetupFrontend(r)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/health", controllers.HealthCheck)
			v1.POST("/auth/login", controllers.Login)
			v1.GET("/system/info", middleware.Auth(), controllers.GetSystemInfo)
			v1.GET("/system/current", middleware.Auth(), controllers.GetCurrentInfo)
			v1.GET("/system/os", middleware.Auth(), controllers.GetOSInfo)
			v1.GET("/config", middleware.Auth(), controllers.GetConfig)
			v1.POST("/config", middleware.Auth(), controllers.UpdateConfig)
			v1.GET("/config/initialized", middleware.Auth(), controllers.CheckConfigInitialized)
			v1.POST("/server/restart", middleware.Auth(), controllers.RestartServer)

			// 系统设置 API
			settingController := controllers.NewSettingController()
			settings := v1.Group("/settings")
			{
				settings.GET("", middleware.Auth(), settingController.GetAllSettings)
				settings.GET("/system", middleware.Auth(), settingController.GetSystemSettings)
				settings.POST("/system", middleware.Auth(), settingController.UpdateSystemSettings)
				settings.GET("/:key", middleware.Auth(), settingController.GetSettingByKey)
				settings.POST("", middleware.Auth(), settingController.CreateSetting)
				settings.PUT("", middleware.Auth(), settingController.UpdateSetting)
				settings.DELETE("/:key", middleware.Auth(), settingController.DeleteSetting)
				settings.GET("/terminal", middleware.Auth(), settingController.GetTerminalInfo)
				settings.POST("/terminal", middleware.Auth(), settingController.UpdateTerminal)
			}

			// 配置热重载 API
			v1.POST("/config/reload", middleware.Auth(), controllers.ReloadConfig)

			// 文件管理 API
			fileController := controllers.NewFileController()
			files := v1.Group("/files")
			{
				files.GET("/drives", middleware.Auth(), fileController.GetDrives)
				files.POST("/list", middleware.Auth(), fileController.ListFiles)
				files.POST("/create", middleware.Auth(), fileController.CreateFile)
				files.POST("/delete", middleware.Auth(), fileController.DeleteFile)
				files.POST("/rename", middleware.Auth(), fileController.RenameFile)
				files.POST("/move", middleware.Auth(), fileController.MoveFile)
				files.POST("/content", middleware.Auth(), fileController.GetFileContent)
				files.POST("/save", middleware.Auth(), fileController.SaveFileContent)
				files.POST("/upload", middleware.Auth(), fileController.UploadFile)
				files.GET("/download", middleware.Auth(), fileController.DownloadFile)
				files.POST("/size", middleware.Auth(), fileController.GetDirSize)
				files.POST("/compress", middleware.Auth(), fileController.CompressFiles)
				files.POST("/decompress", middleware.Auth(), fileController.DecompressFile)
				files.POST("/chmod", middleware.Auth(), fileController.ChmodFile)
				files.POST("/chown", middleware.Auth(), fileController.ChownFile)
				files.GET("/progress", middleware.Auth(), fileController.GetProgress)
				files.POST("/preview", middleware.Auth(), fileController.PreviewFile)
			}

			// WebSocket 连接
			v1.GET("/ws", utils.HandleWebSocket)

			// 终端 WebSocket 代理（代理到 Agent）
			terminalController, _ := controllers.NewTerminalController()
			terminal := v1.Group("/terminal")
			{
				terminal.GET("/local", middleware.Auth(), terminalController.TerminalLocal)
				terminal.GET("/ssh", middleware.Auth(), terminalController.TerminalSSH)
			}

			// 防火墙 API（代理到 Agent）
			firewallController, _ := controllers.NewFirewallController()
			agent := v1.Group("/agent")
			{
				agent.POST("/firewall/base", middleware.Auth(), firewallController.LoadBaseInfo)
				agent.POST("/firewall/search", middleware.Auth(), firewallController.SearchRules)
				agent.POST("/firewall/operate", middleware.Auth(), firewallController.OperateFirewall)
				agent.POST("/firewall/port", middleware.Auth(), firewallController.OperatePortRule)
				agent.POST("/firewall/ip", middleware.Auth(), firewallController.OperateIPRule)
				agent.POST("/firewall/update/port", middleware.Auth(), firewallController.UpdatePortRule)
				agent.POST("/firewall/update/ip", middleware.Auth(), firewallController.UpdateIPRule)
				agent.POST("/firewall/forward", middleware.Auth(), firewallController.OperateForwardRule)
				agent.GET("/firewall/install", middleware.Auth(), firewallController.InstallFirewall)
			}

			// 主机管理 API
			controllers.InitHostController()
			hostGroups := v1.Group("/host-groups")
			{
				hostGroups.GET("", middleware.Auth(), controllers.ListGroups)
				hostGroups.POST("", middleware.Auth(), controllers.CreateGroup)
				hostGroups.GET("/:id", middleware.Auth(), controllers.GetGroupByID)
				hostGroups.PUT("/:id", middleware.Auth(), controllers.UpdateGroup)
				hostGroups.DELETE("/:id", middleware.Auth(), controllers.DeleteGroup)
			}
			hosts := v1.Group("/hosts")
			{
				hosts.GET("", middleware.Auth(), controllers.ListHosts)
				hosts.POST("", middleware.Auth(), controllers.CreateHost)
				hosts.GET("/tree", middleware.Auth(), controllers.GetHostTree)
				hosts.POST("/test", middleware.Auth(), controllers.TestHostConnection)
				hosts.POST("/move", middleware.Auth(), controllers.MoveHosts)
				hosts.GET("/export", middleware.Auth(), controllers.ExportHosts)
				hosts.POST("/import", middleware.Auth(), controllers.ImportHosts)
				hosts.GET("/:id", middleware.Auth(), controllers.GetHostByID)
				hosts.PUT("/:id", middleware.Auth(), controllers.UpdateHost)
				hosts.DELETE("/:id", middleware.Auth(), controllers.DeleteHost)
				hosts.GET("/:id/connection", middleware.Auth(), controllers.GetHostForTerminal)
			}

			// 会话历史 API
			controllers.InitSessionHistoryController()
			sessionHistories := v1.Group("/session-histories")
			{
				sessionHistories.GET("", middleware.Auth(), controllers.ListSessionHistories)
				sessionHistories.POST("", middleware.Auth(), controllers.CreateSessionHistory)
				sessionHistories.POST("/search", middleware.Auth(), controllers.SearchSessionHistories)
				sessionHistories.GET("/:id", middleware.Auth(), controllers.GetSessionHistoryByID)
				sessionHistories.PUT("", middleware.Auth(), controllers.UpdateSessionHistory)
				sessionHistories.DELETE("/:id", middleware.Auth(), controllers.DeleteSessionHistory)
			}

			// 快速命令 API
			quickCommandController := controllers.NewQuickCommandController()
			quickCommands := v1.Group("/quick-commands")
			{
				quickCommands.POST("", middleware.Auth(), quickCommandController.Create)
				quickCommands.POST("/update", middleware.Auth(), quickCommandController.Update)
				quickCommands.POST("/delete", middleware.Auth(), quickCommandController.Delete)
				quickCommands.POST("/search", middleware.Auth(), quickCommandController.Search)
				quickCommands.GET("/all", middleware.Auth(), quickCommandController.GetAll)
			}
		}
	}
}

// SetupFrontend 设置前端静态文件服务
func SetupFrontend(r *gin.Engine) {
	frontendDist, err := fs.Sub(frontendFS, "web/dist")
	if err != nil {
		log.Printf("Warning: Failed to load frontend files: %v", err)
		return
	}

	fileServer := http.FileServer(http.FS(frontendDist))

	// 处理静态资源
	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.Request.URL.Path = "/assets/" + filepath
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// 使用 NoRoute 处理所有其他路由（包括前端路由）
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是 API 请求，返回 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// 对于其他所有请求，返回 index.html 以支持前端路由
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}