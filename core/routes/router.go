package routes

import (
	"gpanel/controllers"
	"gpanel/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/health", controllers.HealthCheck)
			v1.POST("/auth/login", controllers.Login)
			v1.GET("/system/info", middleware.Auth(), controllers.GetSystemInfo)
			v1.GET("/system/current", middleware.Auth(), controllers.GetCurrentInfo)
			v1.GET("/system/version", middleware.Auth(), controllers.GetVersion)
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
			}

			// 配置热重载 API
			v1.POST("/config/reload", middleware.Auth(), controllers.ReloadConfig)
		}
	}
}