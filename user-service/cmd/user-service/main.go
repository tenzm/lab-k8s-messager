package main

import (
	"log"
	"net/http"
	"user-service/internal/config"
	httpctrl "user-service/internal/controller/http"
	"user-service/internal/repository/users"
	"user-service/internal/usecase/getuser"
	"user-service/internal/usecase/register"
	"user-service/internal/usecase/search"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DBDSN)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	repo := users.New(db)

	registerUC := register.NewUseCase(repo)
	searchUC := search.NewUseCase(repo)
	getUserUC := getuser.NewUseCase(repo)

	handler := httpctrl.NewHandler(registerUC, searchUC, getUserUC)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	handler.Register(r)

	log.Printf("user-service listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("run: %v", err)
	}
}
