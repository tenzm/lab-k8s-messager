package main

import (
	"bff/internal/config"
	httpctrl "bff/internal/controller/http"
	"bff/internal/external/messageservice"
	"bff/internal/external/userservice"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	usersClient := userservice.New(cfg.UserServiceURL)
	msgClient := messageservice.New(cfg.MsgServiceURL)

	handler := httpctrl.NewHandler(usersClient, msgClient, cfg.MsgServiceURL)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// serve frontend static files
	r.Static("/static", "./frontend")
	r.StaticFile("/", "./frontend/index.html")
	handler.Register(r)

	log.Printf("bff listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("run: %v", err)
	}
}
