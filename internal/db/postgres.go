package db

import (
	"context"
	"fmt"
	"log"

	"l0/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	SaveOrder(ctx context.Context, order models.Order) error
	GetOrder(ctx context.Context, orderUID string) (*models.Order, error)
	Close()
	GetPool() *pgxpool.Pool
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, connString string) (Database, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return &Postgres{pool: pool}, nil
}

// SaveOrder сохраняет заказ в БД (включая delivery, payment и items).
func (p *Postgres) SaveOrder(ctx context.Context, order models.Order) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SMID, order.DateCreated, order.OOFShard,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into orders: %v", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into delivery: %v", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into payment: %v", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert into items: %v", err)
		}
	}

	return tx.Commit(ctx)
}

// GetOrder возвращает заказ по order_uid.
func (p *Postgres) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	var order models.Order

	// Основные данные заказа
	err := p.pool.QueryRow(ctx, `
        SELECT 
            order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = $1`, orderUID).
		Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %v", err)
	}

	// Данные доставки
	err = p.pool.QueryRow(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1`, orderUID).
		Scan(
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %v", err)
	}

	// Данные оплаты
	err = p.pool.QueryRow(ctx, `
		SELECT 
			transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE order_uid = $1`, orderUID).
		Scan(
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
			&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %v", err)
	}

	// Товары
	rows, err := p.pool.Query(ctx, `
		SELECT 
			chrt_id, track_number, price, rid, name, sale,
			size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %v", err)
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) GetPool() *pgxpool.Pool {
	return p.pool
}
