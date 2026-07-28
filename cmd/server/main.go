package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"national-defense-knowledge-quiz/internal/config"
	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/handler"
	"national-defense-knowledge-quiz/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	db.Init(cfg.DBURL)

	examSvc := service.NewExamService(cfg.TZ)
	examHandler := handler.NewExamHandler(examSvc)

	r := gin.Default()

	r.GET("/", examHandler.ServeIndex)
	r.GET("/exam/info/", examHandler.Info)
	r.POST("/exam/start/", examHandler.Start)
	r.POST("/exam/submit/", examHandler.Submit)

	log.Printf("server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
