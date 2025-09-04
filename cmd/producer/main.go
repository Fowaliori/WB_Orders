package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

type Order struct {
	OrderUID          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerID        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	ShardKey          string   `json:"shardkey"`
	SMID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OOFShard          string   `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NMID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func main() {

	topic := "orders"
	broker := "localhost:9092"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	for {
		order := generateOrder()
		message, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order: %v", err)
			continue
		}

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(order.OrderUID),
				Value: message,
			},
		)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		} else {
			fmt.Printf("Sent order: %s\n", order.OrderUID)
		}

		time.Sleep(3 * time.Second)
	}
}

// Генератор тестовых заказов
func generateOrder() Order {
	orderUID := fmt.Sprintf("test-order-%d", rand.Intn(10000))
	return Order{
		OrderUID:    orderUID,
		TrackNumber: fmt.Sprintf("WBIL-%d", rand.Intn(1000)),
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    fmt.Sprintf("User %d", rand.Intn(100)),
			Phone:   "+" + fmt.Sprintf("%010d", rand.Intn(1000000000)),
			Zip:     fmt.Sprintf("%d", rand.Intn(100000)),
			City:    "Moscow",
			Address: fmt.Sprintf("Street %d", rand.Intn(100)),
			Region:  "Region",
			Email:   fmt.Sprintf("user%d@test.com", rand.Intn(100)),
		},
		Payment: Payment{
			Transaction:  orderUID,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       rand.Intn(10000),
			PaymentDT:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   rand.Intn(500),
			CustomFee:    0,
		},
		Items: []Item{
			{
				ChrtID:      rand.Intn(1000000),
				TrackNumber: fmt.Sprintf("WBIL-%d", rand.Intn(1000)),
				Price:       rand.Intn(1000),
				RID:         fmt.Sprintf("rid-%d", rand.Intn(1000)),
				Name:        "Test Item",
				Sale:        30,
				Size:        "0",
				TotalPrice:  rand.Intn(500),
				NMID:        rand.Intn(1000000),
				Brand:       "Brand",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("customer-%d", rand.Intn(100)),
		DeliveryService:   "meest",
		ShardKey:          "9",
		SMID:              rand.Intn(100),
		DateCreated:       time.Now().Format(time.RFC3339),
		OOFShard:          "1",
	}
}
