package main

import (
	"gpanel/config"
	"gpanel/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gin.SetMode(config.AppConfig.Server.Mode)

	r := gin.Default()

	SetupFrontend(r)
	routes.SetupRouter(r)

	addr := ":" + config.AppConfig.Server.Port
	log.Printf("Starting GPanel server on %s", addr)
	r.Run(addr)
}