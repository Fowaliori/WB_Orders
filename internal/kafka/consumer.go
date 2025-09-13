package kafka

import (
	"context"
	"encoding/json"
	"log"

	"l0/internal/cache"
	"l0/internal/db"
	"l0/internal/models"
	"l0/internal/utils"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	StartConsumer(ctx context.Context, brokers []string, topic string, db db.Database, cacheService cache.Cache) error
}

type KafkaConsumer struct{}

func (k *KafkaConsumer) StartConsumer(ctx context.Context, brokers []string, topic string, dbService db.Database, cacheService cache.Cache) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order-processor",
	})

	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

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

			if err := utils.ValidateStruct(order); err != nil {
				log.Printf("Invalid order data: %v", err)
				continue
			}

			// Сохраняем в БД
			if err := dbService.SaveOrder(ctx, order); err != nil {
				log.Printf("Failed to save order to DB: %v", err)
				continue
			}

			// Сохраняем в кэш
			cacheService.Set(order.OrderUID, order)

			log.Printf("Processed order: %s", order.OrderUID)
		}
	}
}

func StartConsumer(ctx context.Context, brokers []string, topic string, db db.Database, cache cache.Cache) error {
	consumer := &KafkaConsumer{}
	return consumer.StartConsumer(ctx, brokers, topic, db, cache)
}
