package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/user/streaming/internal/analytics"
	"github.com/user/streaming/internal/config"
	"github.com/user/streaming/internal/domain"
	rstream "github.com/user/streaming/internal/redis"
	"github.com/user/streaming/internal/websocket"
	"github.com/user/streaming/internal/worker"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	// 1. Initialize Redis Client
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	// 2. Initialize Components
	stream := rstream.NewRedisStream(rdb, cfg.StreamName)
	engine := analytics.NewEngine(cfg.WindowSize)
	hub := websocket.NewHub()
	pool := worker.NewPool(stream, engine, cfg.WorkerCount)

	// 3. Start Processes
	go hub.Run()
	pool.Start(ctx)

	// 4. Start Event Generator (Simulates High-Frequency Ingestion)
	const numGenerators = 5
	const targetEPS = 1500 // Target 1500 events per second
	interval := time.Second / time.Duration(targetEPS/numGenerators)

	for i := 0; i < numGenerators; i++ {
		go func(genID int) {
			types := []string{"order_created", "user_login", "page_view", "payment_success"}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for range ticker.C {
				event := domain.Event{
					ID:        fmt.Sprintf("evt-%d-%d", genID, time.Now().UnixNano()),
					Type:      types[rand.Intn(len(types))],
					UserID:    fmt.Sprintf("user-%d", rand.Intn(100)),
					Value:     rand.Float64() * 100,
					Timestamp: time.Now(),
				}
				stream.Publish(ctx, event)
			}
		}(i)
	}

	// 5. Start Metric Broadcaster (Pushes data to Dashboard via WS)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			metrics := engine.GetLiveMetrics()
			hub.Broadcast(metrics)
		}
	}()

	// 6. Setup API Routes
	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		hub.HandleWebSocket(c.Writer, c.Request)
	})

	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, engine.GetLiveMetrics())
	})

	log.Println("Server starting on :8080")
	r.Run(":8080")
}