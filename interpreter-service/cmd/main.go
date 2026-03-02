package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/handler"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/infrastructure/cache"
	configinfra "github.com/Oleja123/code-vizualization/interpreter-service/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to YAML config")
	flag.Parse()

	cfg := configinfra.LoadOrDefault(*configPath)

	listenPort := cfg.ServerConfig.Port

	var cacher cache.Cacher
	if cfg.RedisConfig.Host != "" {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.RedisConfig.Host, cfg.RedisConfig.Port),
			Password: cfg.RedisConfig.Password,
			DB:       cfg.RedisConfig.DB,
		})

		for attempt := 1; attempt <= cfg.PingAttempts; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := redisClient.Ping(ctx).Err()
			cancel()

			if err == nil {
				cacher = cache.NewRedisCacher(redisClient, cfg.RedisConfig.Expiration)
				break
			}

			if attempt == cfg.PingAttempts {
				log.Printf("redis disabled after %d attempts: %v", cfg.PingAttempts, err)
				cacher = nil
				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	http.Handle("/snapshot", handler.NewSnapshotHandler(*configPath, cacher))

	address := fmt.Sprintf(":%d", listenPort)
	log.Printf("interpreter-service listening on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
