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
			v1.GET("/config", middleware.Auth(), controllers.GetConfig)
			v1.POST("/config", middleware.Auth(), controllers.UpdateConfig)
			v1.GET("/config/initialized", middleware.Auth(), controllers.CheckConfigInitialized)
			v1.POST("/server/restart", middleware.Auth(), controllers.RestartServer)
		}
	}
}