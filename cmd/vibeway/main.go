package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"vibeway/internal/config"
	"vibeway/internal/router"
	"vibeway/internal/tracing"
	"vibeway/internal/upstream"
	"vibeway/pkg/cache"
	"vibeway/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func main() {
	// 1. Load Config
	if err := config.LoadConfig("configs/routes.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init Logger
	logger.InitLogger(config.AppConfig.Server.Mode)

	// 3. Init Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if err := cache.InitRedis(redisAddr, "", 0); err != nil {
		logger.Error("Failed to init redis", err, nil)
		// Continue or fatal? Fatal for production
	}

	// 4. Init Tracing
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4317"
	}
	// Strip http:// or https:// if present for gRPC endpoint if needed, 
    // but InitTracer likely expects host:port. 
    // The docker-compose has http://jaeger:4317 which might be for HTTP exporter.
    // Let's check InitTracer implementation. 
    // Assuming InitTracer takes a raw endpoint string.
    // If InitTracer uses grpc.WithEndpoint, it expects host:port.
    // Let's clean it up just in case.
    if len(otelEndpoint) > 7 && (otelEndpoint[:7] == "http://" || otelEndpoint[:8] == "https://") {
        // Simple strip for now, or just trust the env var matches what InitTracer expects.
        // Let's stick to the env var value for now.
    }

	tp, err := tracing.InitTracer(context.Background(), "vibeway", otelEndpoint)
	if err != nil {
		logger.Error("Failed to init tracing", err, nil)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider", err, nil)
		}
	}()

	// 5. Init Upstream Manager
	upstreams := upstream.NewManager(config.AppConfig.Upstreams)

	// 6. Init Fiber
	app := fiber.New(fiber.Config{
		AppName: "Vibeway",
	})

	// 7. Setup Routes
	router.SetupRoutes(app, config.AppConfig, upstreams)

	// 8. Metrics Endpoint
	// 8. Metrics Endpoint
	app.Get("/metrics", func(c fiber.Ctx) error {
		handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
		handler(c.RequestCtx())
		return nil
	})

	// 9. Health Check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// 10. Start Server
	go func() {
		addr := ":" + strconv.Itoa(config.AppConfig.Server.Port)
		if err := app.Listen(addr); err != nil {
			logger.Error("Server failed to start", err, nil)
		}
	}()

	// Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down server...", nil)
	_ = app.Shutdown()
}
