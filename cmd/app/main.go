package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"l0/internal/api"
	"l0/internal/cache"
	"l0/internal/db"
	"l0/internal/kafka"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подключение к PostgreSQL
	pg, err := db.NewPostgres(ctx, "postgres://user_wb:123@localhost:5432/level0?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pg.Close()

	// Инициализация кэша
	cache := cache.NewCache()
	if err := restoreCacheFromDB(ctx, pg, cache); err != nil {
		log.Printf("Failed to restore cache from DB: %v", err)
	}

	// Запуск Kafka-консьюмера
	go func() {
		if err := kafka.StartConsumer(
			ctx,
			[]string{"localhost:9092"},
			"orders",
			pg,
			cache,
		); err != nil {
			log.Fatalf("Kafka consumer failed: %v", err)
		}
	}()

	// Настройка HTTP-сервера
	handler := api.NewHandler(cache)
	server := &http.Server{
		Addr:    ":8082",
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		cancel()
	}()

	log.Println("Server starting on :8082...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
}

func restoreCacheFromDB(ctx context.Context, pg *db.Postgres, cache *cache.Cache) error {
	// Здесь реализация восстановления кэша из БД
	return nil
}
