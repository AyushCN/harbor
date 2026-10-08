package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/yourusername/harbor/internal/config"
	"github.com/yourusername/harbor/internal/database"
	harborhttp "github.com/yourusername/harbor/internal/http"
	"github.com/yourusername/harbor/internal/presence"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	ctx := context.Background()

	// Initialize database
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close(pool)

	// Run migrations
	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Initialize Redis
	redisClient, err := presence.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer presence.CloseRedisClient(redisClient)

	// Initialize presence manager
	pm := presence.NewManager(redisClient)
	defer pm.Close()

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Create router
	router := harborhttp.NewRouter(cfg, pool, pm, nc)

	addr := ":" + cfg.Port

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Harbor API starting on %s (env=%s)", addr, cfg.Env)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}

	log.Println("Server stopped")
}