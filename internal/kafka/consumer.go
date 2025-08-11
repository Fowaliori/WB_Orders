package kafka

import (
	"context"
	"encoding/json"
	"log"

	"l0/internal/cache"
	"l0/internal/db"
	"l0/internal/models"

	"github.com/segmentio/kafka-go"
)

func StartConsumer(ctx context.Context, brokers []string, topic string, pg *db.Postgres, cache *cache.Cache) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order-processor",
	})

	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				return err
			}

			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Failed to unmarshal order: %v", err)
				continue
			}

			// Сохраняем в БД
			if err := pg.SaveOrder(ctx, order); err != nil {
				log.Printf("Failed to save order to DB: %v", err)
				continue
			}

			// Сохраняем в кэш
			cache.Set(order.OrderUID, order)

			log.Printf("Processed order: %s", order.OrderUID)
		}
	}
}
