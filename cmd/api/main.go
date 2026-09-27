package main

import (
	"context"
	"enginer/internal/repository/postgres"
	"enginer/internal/service"
	httptransport "enginer/internal/transport/http_transport"
	"fmt"
	"log"
	"os"

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
	ctx := context.Background()
	pool, err := postgres.NewDB(ctx, dsn)
	if err != nil {
		log.Fatal("failed to connect to db: ", err)
	}
	defer pool.Close()

	projectRepo := postgres.NewProjectRepo(pool)
	systemRepo := postgres.NewSystemRepo(pool)
	segmentRepo := postgres.NewSegmentRepo(pool)

	segmentSvc := service.NewSegmentService(segmentRepo, systemRepo)
	systemSvc := service.NewSystemService(systemRepo, projectRepo)
	projectHandlers := httptransport.NewProjectHandlers(projectRepo)
	systemHadlers := httptransport.NewSystemHandlers(systemSvc)
	segmentHandlers := httptransport.NewSegmentHandlers(segmentSvc)
	httpServer := httptransport.NewHTTPServer(projectHandlers, systemHadlers, segmentHandlers)

	if err := httpServer.StartServer(); err != nil {
		fmt.Println("failed to start http server", err)
	}
}
