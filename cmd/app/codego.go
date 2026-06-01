package go

import (
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"myapp/internal/config"
	"myapp/internal/handler"
	"myapp/internal/middleware"
	"myapp/internal/redis"
	"myapp/internal/service"
)

func main() {
  cfg, err := config.Load("configs/main.yaml")
  if err != nil {
    log.Fatalf("load config: %v", err)
  }

  redisClient, err := repository.InitRedis(cfg.Redis)
  if err != nil {
    log.Fatalf("redis connect: %v", err)
  }
  defer redisClient.Close()

  sessionService := service.NewSessionService(redisClient)
  pingService := service.NewPingService(redisClient)

  authHandler := handler.NewAuthHandler(sessionService)
  pingHandler := handler.NewPingHandler(pingService)

  router := gin.Default()
  router.POST("/login", authHandler.Login)

  protected := router.Group("/")
  protected.Use(middleware.Auth(sessionService))
  protected.GET("/ping", pingHandler.Ping)
  protected.GET("/top", pingHandler.Top)
  protected.GET("/count", pingHandler.Count)

  addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
  if err := router.Run(addr); err != nil {
    log.Fatalf("server error: %v", err)
  }
}