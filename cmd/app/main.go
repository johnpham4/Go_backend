package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"myapp/internal/config"
	repository "myapp/internal/redis"
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

	router := gin.Default()
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}