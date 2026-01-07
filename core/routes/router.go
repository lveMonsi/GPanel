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
		}
	}
}