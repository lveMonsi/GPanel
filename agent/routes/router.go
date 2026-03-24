package routes

import (
	"gpanel/agent/controllers"
	"gpanel/agent/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 应用安全中间件
	r.Use(middleware.Security())

	// API 路由（需要 API Key 认证）
	api := r.Group("/api")
	api.Use(middleware.Auth())
	{
		v1 := api.Group("/v1")
		{
			// 防火墙路由
			firewallController, _ := controllers.NewFirewallController()
			firewall := v1.Group("/firewall")
			{
				firewall.POST("/base", firewallController.LoadBaseInfo)
				firewall.POST("/search", firewallController.SearchRules)
				firewall.POST("/operate", firewallController.OperateFirewall)
				firewall.POST("/port", firewallController.OperatePortRule)
				firewall.POST("/update/port", firewallController.UpdatePortRule)
				firewall.POST("/ip", firewallController.OperateIPRule)
				firewall.POST("/update/ip", firewallController.UpdateIPRule)
				firewall.POST("/forward", firewallController.OperateForwardRule)
				firewall.GET("/install", firewallController.InstallFirewall)
				firewall.GET("/uninstall", firewallController.UninstallFirewall)
			}

			// 终端路由
			terminalController, _ := controllers.NewTerminalController()
			terminal := v1.Group("/terminal")
			{
				terminal.GET("/local", terminalController.TerminalLocal)
				terminal.GET("/ssh", terminalController.TerminalSSH)
			}

			// 主机路由
			hostController, _ := controllers.NewHostController()
			hosts := v1.Group("/hosts")
			{
				hosts.POST("/test", hostController.TestConnection)
			}

			// 系统信息路由
			systemController, _ := controllers.NewSystemController()
			system := v1.Group("/system")
			{
				system.GET("/info", systemController.GetSystemInfo)
				system.GET("/current", systemController.GetCurrentInfo)
				system.GET("/os", systemController.GetOSInfo)
			}

			// 文件管理路由
			fileController := controllers.NewFileController()
			files := v1.Group("/files")
			{
				files.GET("/drives", fileController.GetDrives)
				files.POST("/list", fileController.ListFiles)
				files.POST("/create", fileController.CreateFile)
				files.POST("/delete", fileController.DeleteFile)
				files.POST("/rename", fileController.RenameFile)
				files.POST("/move", fileController.MoveFile)
				files.POST("/content", fileController.GetFileContent)
				files.POST("/save", fileController.SaveFileContent)
				files.POST("/upload", fileController.UploadFile)
				files.GET("/download", fileController.DownloadFile)
				files.POST("/size", fileController.GetDirSize)
				files.POST("/compress", fileController.CompressFiles)
				files.POST("/decompress", fileController.DecompressFile)
				files.POST("/chmod", fileController.ChmodFile)
				files.POST("/chown", fileController.ChownFile)
				files.GET("/progress", fileController.GetProgress)
				files.POST("/preview", fileController.PreviewFile)
				files.POST("/remote-download", fileController.RemoteDownload)
			}

			// 设置路由
			settingController, _ := controllers.NewSettingController()
			settings := v1.Group("/settings")
			{
				settings.POST("/ssh", settingController.SaveLocalConn)
				settings.GET("/ssh/conn", settingController.GetLocalConn)
				settings.POST("/ssh/check", settingController.TestLocalConn)
			}
		}
	}
}
