package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Восстановление кэша из БД
	if err := restoreCacheFromDB(ctx, pg, cache); err != nil {
		log.Printf("Failed to restore cache from DB: %v", err)
	} else {
		log.Printf("Cache restored successfully. Total orders in cache: %d", len(cache.GetAll()))
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
	handler := api.NewHandler(cache, pg)
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
	start := time.Now()

	// Получаем 30 последних order_uid из таблицы orders
	rows, err := pg.GetPool().Query(ctx, `
        SELECT order_uid 
        FROM orders 
        ORDER BY date_created DESC 
        LIMIT 30`)
	if err != nil {
		return fmt.Errorf("failed to query orders: %v", err)
	}
	defer rows.Close()

	var orderUIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return fmt.Errorf("failed to scan order_uid: %v", err)
		}
		orderUIDs = append(orderUIDs, uid)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %v", err)
	}

	// Для каждого order_uid загружаем полный заказ
	for _, uid := range orderUIDs {
		order, err := pg.GetOrder(ctx, uid)
		if err != nil {
			log.Printf("Failed to load order %s: %v", uid, err)
			continue
		}
		cache.Set(uid, *order)
	}

	log.Printf("Cache restoration completed. Loaded %d recent orders in %v", len(orderUIDs), time.Since(start))
	return nil
}
