package main

import (
	"context"
	"enginer/internal/repository/postgres"
	redisstore "enginer/internal/repository/redis_store"
	"enginer/internal/service"
	httptransport "enginer/internal/transport/http_transport"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println(".env не найден, читем переменные окружения напрямгую")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := postgres.NewDB(ctx, dsn)
	if err != nil {
		log.Fatal("failed to connect to db: ", err)
	}
	defer pool.Close()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		log.Fatalf("REDIS_ADDR is required")
	}
	client, err := redisstore.NewClient(ctx, addr)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer client.Close()

	projectRepo := postgres.NewProjectRepo(pool)
	systemRepo := postgres.NewSystemRepo(pool)
	segmentRepo := postgres.NewSegmentRepo(pool)

	segmentSvc := service.NewSegmentService(segmentRepo, systemRepo, client)
	systemSvc := service.NewSystemService(systemRepo, projectRepo)
	projectHandlers := httptransport.NewProjectHandlers(projectRepo)
	systemHadlers := httptransport.NewSystemHandlers(systemSvc)
	segmentHandlers := httptransport.NewSegmentHandlers(segmentSvc)
	httpServer := httptransport.NewHTTPServer(projectHandlers, systemHadlers, segmentHandlers)

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":9091"
	}
	log.Printf("http server listening on %s", httpAddr)
	if err := httpServer.StartServer(ctx, httpAddr); err != nil {
		fmt.Println("failed to start http server", err)
	}
}
