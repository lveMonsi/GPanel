package main

import (
	"gpanel/config"
	"gpanel/middleware"
	"gpanel/routes"
	"log"
	"runtime"

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

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gin.SetMode(config.AppConfig.Server.Mode)

	r := gin.Default()

	// 添加安全入口中间件
	r.Use(middleware.SecurityEntrance())

	// 添加安全中间件
	r.Use(middleware.Security())

	SetupFrontend(r)
	routes.SetupRouter(r)

	addr := ":" + config.AppConfig.Server.Port
	log.Printf("Starting GPanel server on %s", addr)
	log.Printf("Security entrance: %s", config.AppConfig.SecurityEntrance)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}