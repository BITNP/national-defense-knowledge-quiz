package main

import (
	"context"
	"log"
	"net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"national-defense-knowledge-quiz/internal/cache"
	"national-defense-knowledge-quiz/internal/config"
	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/handler"
	"national-defense-knowledge-quiz/internal/middleware"
	"national-defense-knowledge-quiz/internal/repository"
	"national-defense-knowledge-quiz/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	db.Init(cfg.DBType, cfg.DBURL)

	problemCache := cache.NewProblemCache()
	if err := problemCache.LoadAll(context.Background(), &repository.ProblemRepo{}); err != nil {
		log.Fatalf("failed to warm problem cache: %v", err)
	}

	examCache := cache.NewExamCache()
	if err := examCache.LoadAll(context.Background(), &repository.ExamRepo{}); err != nil {
		log.Fatalf("failed to warm exam cache: %v", err)
	}

	examSvc := service.NewExamService(cfg.TZ, problemCache, examCache)
	examHandler := handler.NewExamHandler(examSvc)

	r := gin.Default()
	r.Use(middleware.RequestLatency())
	r.Use(middleware.RequestTimeout(10 * time.Second))

	r.GET("/", examHandler.ServeIndex)
	r.GET("/exam/info/", examHandler.Info)
	r.POST("/exam/start/", examHandler.Start)
	r.POST("/exam/submit/", examHandler.Submit)

	// Profiling endpoints (local/dev only; do not expose publicly in production).
	debug := r.Group("/debug/pprof")
	debug.GET("/", gin.WrapF(pprof.Index))
	debug.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	debug.GET("/profile", gin.WrapF(pprof.Profile))
	debug.GET("/symbol", gin.WrapF(pprof.Symbol))
	debug.POST("/symbol", gin.WrapF(pprof.Symbol))
	debug.GET("/trace", gin.WrapF(pprof.Trace))
	debug.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
	debug.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
	debug.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
	debug.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
	debug.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
	debug.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))

	log.Printf("server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
