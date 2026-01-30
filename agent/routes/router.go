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

	// API 路由
	api := r.Group("/api")
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
			}
		}
	}
}