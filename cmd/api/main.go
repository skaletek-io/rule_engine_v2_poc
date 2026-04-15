package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	apidocs "github.com/skaletek/rule-engine-v2-poc/api"
	api "github.com/skaletek/rule-engine-v2-poc/internal/api"
	"github.com/skaletek/rule-engine-v2-poc/internal/api/handlers"
	platformdb "github.com/skaletek/rule-engine-v2-poc/internal/platform/db"
	"github.com/skaletek/rule-engine-v2-poc/internal/platform/db/store"
)

func main() {
	// Load .env if present; production environments inject vars directly.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: could not load .env: %v", err)
	}

	ctx := context.Background()

	pool, err := platformdb.NewPool(ctx)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := platformdb.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	s := store.New(pool)
	h := handlers.New(s)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Mount all generated routes under /api/v1
	api.RegisterHandlersWithBaseURL(e, api.NewStrictHandler(h, nil), "/api/v1")

	// Swagger UI + raw spec at /api/v1/docs and /api/v1/openapi.yaml
	apidocs.RegisterDocs(e.Group("/api/v1"), "/api/v1/openapi.yaml")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("starting server on :%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
