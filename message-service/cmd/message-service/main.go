package main

import (
	"log"
	"message-service/internal/config"
	httpctrl "message-service/internal/controller/http"
	"message-service/internal/repository/files"
	"message-service/internal/repository/messages"
	"message-service/internal/usecase/deletemessage"
	"message-service/internal/usecase/editmessage"
	"message-service/internal/usecase/getconversations"
	"message-service/internal/usecase/getmessages"
	"message-service/internal/usecase/sendmessage"
	"message-service/internal/usecase/uploadfile"
	"net/http"

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

	msgRepo := messages.New(db)
	fileRepo := files.New(db)

	sendUC := sendmessage.NewUseCase(msgRepo)
	editUC := editmessage.NewUseCase(msgRepo)
	deleteUC := deletemessage.NewUseCase(msgRepo)
	getUC := getmessages.NewUseCase(msgRepo)
	getConvsUC := getconversations.NewUseCase(msgRepo)
	uploadUC := uploadfile.NewUseCase(fileRepo, cfg.UploadsDir)

	handler := httpctrl.NewHandler(sendUC, editUC, deleteUC, getUC, getConvsUC, uploadUC, cfg.UploadsDir)

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

	log.Printf("message-service listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("run: %v", err)
	}
}
