package main

import (
	"context"
	"errors"
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
	"l0/internal/models"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbService, err := db.NewPostgres(ctx, "postgres://user_wb:123@localhost:5433/level0?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer dbService.Close()

	cacheService := cache.NewCache()

	if err := restoreCacheFromDB(ctx, dbService, cacheService); err != nil {
		log.Printf("Failed to restore cache from DB: %v", err)
	} else {
		log.Printf("Cache restored successfully. Total orders in cache: %d", len(cacheService.GetAll()))
	}

	go func() {
		if err := kafka.StartConsumer(
			ctx,
			[]string{"localhost:9092"},
			"orders",
			dbService,
			cacheService,
		); err != nil {
			log.Fatalf("Kafka consumer failed: %v", err)
		}
	}()

	apiHandler := api.NewHandler(cacheService, dbService)
	server := &http.Server{
		Addr:    ":8082",
		Handler: apiHandler,
	}

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
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Server failed:", err)
	}
}

func restoreCacheFromDB(ctx context.Context, pg db.Database, cacheService cache.Cache) error {
	start := time.Now()

	pool := pg.GetPool()

	rows, err := pool.Query(ctx, `
		SELECT 
			o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
			o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
			i.chrt_id, i.track_number, i.price, i.rid, i.name,
			i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
		FROM orders o
		LEFT JOIN delivery d ON o.order_uid = d.order_uid
		LEFT JOIN payment p ON o.order_uid = p.order_uid
		LEFT JOIN items i ON o.order_uid = i.order_uid
		ORDER BY o.date_created DESC
		LIMIT 30`)
	if err != nil {
		return fmt.Errorf("failed to query orders with joins: %v", err)
	}
	defer rows.Close()

	ordersMap := make(map[string]*models.Order)

	for rows.Next() {
		var (
			o models.Order
			d models.Delivery
			p models.Payment
			i models.Item
		)

		err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
			&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SMID, &o.DateCreated, &o.OOFShard,
			&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
			&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDT,
			&p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
			&i.ChrtID, &i.TrackNumber, &i.Price, &i.RID, &i.Name,
			&i.Sale, &i.Size, &i.TotalPrice, &i.NMID, &i.Brand, &i.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		if existing, ok := ordersMap[o.OrderUID]; ok {
			existing.Items = append(existing.Items, i)
		} else {
			o.Delivery = d
			o.Payment = p
			o.Items = []models.Item{i}
			ordersMap[o.OrderUID] = &o
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %v", err)
	}

	for uid, order := range ordersMap {
		cacheService.Set(uid, *order)
	}

	log.Printf("Cache restoration completed. Loaded %d recent orders in %v", len(ordersMap), time.Since(start))
	return nil
}
